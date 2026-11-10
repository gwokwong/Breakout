<template>
  <div class="task-list">
    <!-- 筛选栏 -->
    <div class="card-box">
      <el-form inline size="small">
        <el-form-item label="项目">
          <el-select v-model="filters.project_id" clearable placeholder="全部项目" style="width: 160px" @change="handleSearch">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 120px" @change="handleSearch">
            <el-option v-for="s in TASK_STATUS_LIST" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="filters.priority" clearable placeholder="全部" style="width: 100px" @change="handleSearch">
            <el-option v-for="p in PRIORITY_LIST" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="负责人">
          <el-select v-model="filters.assignee_id" clearable filterable placeholder="全部" style="width: 120px" @change="handleSearch">
            <el-option v-for="m in allMembers" :key="m.id" :label="m.nickname" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-select v-model="filters.sort_by" style="width: 120px" @change="handleSearch">
            <el-option label="创建时间" value="created_at" />
            <el-option label="截止日期" value="due_date" />
            <el-option label="优先级" value="priority" />
            <el-option label="状态" value="status" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button link :type="filters.sort_order === 'asc' ? 'primary' : ''" @click="toggleSortOrder">
            {{ filters.sort_order === 'asc' ? '升序' : '降序' }}
            <el-icon><Sort /></el-icon>
          </el-button>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="搜索标题" clearable style="width: 160px" @clear="handleSearch" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 批量操作栏 -->
    <div class="card-box" v-if="selectedIds.length">
      <div class="batch-bar">
        <span>已选 {{ selectedIds.length }} 项</span>
        <el-select v-model="batchStatus" placeholder="批量改状态" size="small" style="width: 130px" @change="handleBatchStatus">
          <el-option v-for="s in TASK_STATUS_LIST" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-button type="danger" size="small" @click="handleBatchDelete">批量删除</el-button>
      </div>
    </div>

    <!-- 任务表格 -->
    <div class="card-box">
      <el-table :data="tasks" border size="small" v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="title" label="任务标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.title }}</el-link>
            <el-tag v-if="row.blocked_reason" type="danger" size="small" style="margin-left: 4px">阻塞</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="项目" width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.project?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="TASK_STATUS_MAP[row.status]?.type" size="small">{{ TASK_STATUS_MAP[row.status]?.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="70">
          <template #default="{ row }">
            <el-tag :type="PRIORITY_MAP[row.priority]?.type" size="small">{{ PRIORITY_MAP[row.priority]?.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="90">
          <template #default="{ row }">{{ row.assignee?.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column label="截止日期" width="110">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isOverdue(row.due_date, row.status) }">{{ formatDate(row.due_date) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="110">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadTasks"
          @current-change="loadTasks"
        />
      </div>
    </div>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showDetail" :task-id="currentTaskId" @updated="loadTasks" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, deleteTask, updateTaskStatus } from '@/api/task'
import { getProjects } from '@/api/project'
import { getTeam } from '@/api/team'
import { TASK_STATUS_MAP, TASK_STATUS_LIST, PRIORITY_MAP, PRIORITY_LIST } from '@/constants'
import { formatDate, isOverdue } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const loading = ref(false)
const tasks = ref([])
const projects = ref([])
const allMembers = ref([])
const selectedIds = ref([])
const batchStatus = ref(null)
const showDetail = ref(false)
const currentTaskId = ref(0)

const filters = reactive({
  project_id: null,
  status: null,
  priority: null,
  assignee_id: null,
  keyword: '',
  sort_by: 'created_at',
  sort_order: 'desc',
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
})

function toggleSortOrder() {
  filters.sort_order = filters.sort_order === 'asc' ? 'desc' : 'asc'
  handleSearch()
}

function handleSearch() {
  pagination.page = 1
  loadTasks()
}

function handleReset() {
  filters.project_id = null
  filters.status = null
  filters.priority = null
  filters.assignee_id = null
  filters.keyword = ''
  filters.sort_by = 'created_at'
  filters.sort_order = 'desc'
  handleSearch()
}

function handleSelectionChange(rows) {
  selectedIds.value = rows.map((r) => r.id)
}

function openDetail(row) {
  currentTaskId.value = row.id
  showDetail.value = true
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确定删除任务「${row.title}」吗？`, '警告', { type: 'warning' })
  await deleteTask(row.id)
  ElMessage.success('删除成功')
  await loadTasks()
}

async function handleBatchStatus(status) {
  if (!status || !selectedIds.value.length) return
  await ElMessageBox.confirm(`确定将选中的${selectedIds.value.length}个任务状态改为「${TASK_STATUS_MAP[status]?.label}」吗？`, '提示', { type: 'warning' })
  let ok = 0
  let fail = 0
  for (const id of selectedIds.value) {
    try {
      await updateTaskStatus(id, { status })
      ok++
    } catch {
      fail++
    }
  }
  if (fail) {
    ElMessage.warning(`批量更新完成：成功${ok}个，失败${fail}个`)
  } else {
    ElMessage.success('批量更新成功')
  }
  batchStatus.value = null
  await loadTasks()
}

async function handleBatchDelete() {
  await ElMessageBox.confirm(`确定删除选中的${selectedIds.value.length}个任务吗？`, '警告', { type: 'warning' })
  let ok = 0
  let fail = 0
  for (const id of selectedIds.value) {
    try {
      await deleteTask(id)
      ok++
    } catch {
      fail++
    }
  }
  if (fail) {
    ElMessage.warning(`批量删除完成：成功${ok}个，失败${fail}个`)
  } else {
    ElMessage.success('批量删除成功')
  }
  await loadTasks()
}

async function loadTasks() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      sort_by: filters.sort_by,
      sort_order: filters.sort_order,
    }
    if (filters.project_id) params.project_id = filters.project_id
    if (filters.status) params.status = filters.status
    if (filters.priority) params.priority = filters.priority
    if (filters.assignee_id) params.assignee_id = filters.assignee_id
    if (filters.keyword) params.keyword = filters.keyword

    const res = await getTasks(params)
    tasks.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

async function loadFilters() {
  try {
    const [projRes, teamRes] = await Promise.all([getProjects(), getTeam()])
    projects.value = projRes.data?.list || projRes.data || []
    allMembers.value = teamRes.data?.members || []
  } catch { /* 忽略 */ }
}

onMounted(() => {
  loadFilters()
  loadTasks()
})
</script>

<style lang="scss" scoped>
.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.text-danger {
  color: #f25555;
}
</style>