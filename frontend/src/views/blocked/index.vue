<template>
  <div class="blocked-map">
    <!-- 顶部统计 -->
    <el-row :gutter="16" class="mb-16">
      <el-col :span="6">
        <div class="stat-card stat-card--high">
          <div class="stat-value">{{ highRiskCount }}</div>
          <div class="stat-label">高风险</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card stat-card--mid">
          <div class="stat-value">{{ midRiskCount }}</div>
          <div class="stat-label">中风险</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card stat-card--low">
          <div class="stat-value">{{ lowRiskCount }}</div>
          <div class="stat-label">低风险</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card stat-card--total">
          <div class="stat-value">{{ blockedTasks.length }}</div>
          <div class="stat-label">阻塞总数</div>
        </div>
      </el-col>
    </el-row>

    <!-- 阻塞任务列表 -->
    <div class="card-box">
      <div class="flex-between mb-16">
        <h3 class="page-title">
          <el-icon color="#f25555"><Warning /></el-icon> 阻塞地图
        </h3>
        <div class="toolbar">
          <el-select v-model="projectId" clearable placeholder="全部项目" style="width: 180px" @change="loadData">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-button @click="loadData">
            <el-icon><Refresh /></el-icon>&nbsp;刷新
          </el-button>
        </div>
      </div>

      <el-table :data="sortedTasks" border v-loading="loading" size="small">
        <el-table-column label="风险等级" width="100" sortable :sort-by="row => calcRiskScore(row)">
          <template #default="{ row }">
            <div class="risk-badge" :style="{ background: getRiskLevel(calcRiskScore(row)).color }">
              {{ getRiskLevel(calcRiskScore(row)).label }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="风险分" width="80" sortable :sort-by="row => calcRiskScore(row)">
          <template #default="{ row }">
            <span class="risk-score">{{ calcRiskScore(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="任务名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.title }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="阻塞原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="blocked-reason-text">{{ row.blocked_reason }}</span>
          </template>
        </el-table-column>
        <el-table-column label="阻塞天数" width="90" sortable :sort-by="row => calcBlockedDays(row.blocked_since)">
          <template #default="{ row }">
            <el-tag :type="calcBlockedDays(row.blocked_since) > 3 ? 'danger' : 'warning'" size="small">
              {{ calcBlockedDays(row.blocked_since) }}天
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="90">
          <template #default="{ row }">{{ row.assignee?.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column label="优先级" width="70">
          <template #default="{ row }">
            <el-tag :type="PRIORITY_MAP[row.priority]?.type" size="small">{{ PRIORITY_MAP[row.priority]?.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="截止日期" width="110">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isOverdue(row.due_date, row.status) }">{{ formatDate(row.due_date) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="快捷操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" size="small" @click="handleUnblock(row)">解除阻塞</el-button>
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && !blockedTasks.length" description="暂无阻塞任务，一切正常" />
    </div>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showDetail" :task-id="currentTaskId" @updated="loadData" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBlockedTasks, unblockTask } from '@/api/task'
import { getProjects } from '@/api/project'
import { PRIORITY_MAP } from '@/constants'
import { calcRiskScore, getRiskLevel, calcBlockedDays, formatDate, isOverdue } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const loading = ref(false)
const projectId = ref(null)
const projects = ref([])
const blockedTasks = ref([])
const showDetail = ref(false)
const currentTaskId = ref(0)

// 按风险分降序排序
const sortedTasks = computed(() =>
  [...blockedTasks.value].sort((a, b) => calcRiskScore(b) - calcRiskScore(a))
)

const highRiskCount = computed(() => sortedTasks.value.filter((t) => calcRiskScore(t) >= 60).length)
const midRiskCount = computed(() => sortedTasks.value.filter((t) => {
  const s = calcRiskScore(t)
  return s >= 30 && s < 60
}).length)
const lowRiskCount = computed(() => sortedTasks.value.filter((t) => calcRiskScore(t) < 30).length)

function openDetail(row) {
  currentTaskId.value = row.id
  showDetail.value = true
}

async function handleUnblock(row) {
  await ElMessageBox.confirm(`确定解除「${row.title}」的阻塞吗？`, '提示', { type: 'warning' })
  await unblockTask(row.id)
  ElMessage.success('已解除阻塞')
  await loadData()
}

async function loadData() {
  loading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if (projectId.value) params.project_id = projectId.value
    const res = await getBlockedTasks(params)
    blockedTasks.value = res.data?.list || res.data || []
  } finally {
    loading.value = false
  }
}

async function loadProjects() {
  try {
    const res = await getProjects()
    projects.value = res.data?.list || res.data || []
  } catch { /* 忽略 */ }
}

onMounted(() => {
  loadProjects()
  loadData()
})
</script>

<style lang="scss" scoped>
.page-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.toolbar {
  display: flex;
  gap: 12px;
}

.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);

  .stat-value {
    font-size: 28px;
    font-weight: 700;
  }

  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }

  &--high .stat-value { color: #f25555; }
  &--mid .stat-value { color: #f0a020; }
  &--low .stat-value { color: #31c49c; }
  &--total .stat-value { color: #4d7bfe; }
}

.risk-badge {
  color: #fff;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  text-align: center;
}

.risk-score {
  font-weight: 700;
  font-size: 14px;
}

.blocked-reason-text {
  color: #f25555;
  font-size: 13px;
}

.text-danger {
  color: #f25555;
}
</style>