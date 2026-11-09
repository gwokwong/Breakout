<template>
  <div class="dashboard" v-loading="loading">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6">
        <div class="ad-stat-card">
          <div class="ad-stat-icon" style="background: #eef2ff; color: #4d7bfe">
            <el-icon :size="24"><List /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value">{{ stats.total_tasks || 0 }}</div>
            <div class="ad-stat-label">总任务数</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="ad-stat-card">
          <div class="ad-stat-icon" style="background: #e8f8f2; color: #31c49c">
            <el-icon :size="24"><CircleCheck /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #31c49c">{{ stats.completed_tasks || 0 }}</div>
            <div class="ad-stat-label">已完成</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="ad-stat-card">
          <div class="ad-stat-icon" style="background: #fef0f0; color: #f25555">
            <el-icon :size="24"><Warning /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #f25555">{{ stats.blocked_tasks || 0 }}</div>
            <div class="ad-stat-label">阻塞中</div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="ad-stat-card">
          <div class="ad-stat-icon" style="background: #fdf6ec; color: #f0a020">
            <el-icon :size="24"><TrendCharts /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #f0a020">{{ stats.completion_rate || 0 }}%</div>
            <div class="ad-stat-label">完成率</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 阻塞地图摘要 + 项目概览 -->
    <el-row :gutter="16" class="main-row">
      <el-col :span="14">
        <div class="card-box">
          <div class="flex-between mb-16">
            <h3 class="section-title">
              <el-icon color="#f25555"><Warning /></el-icon> 阻塞地图摘要
            </h3>
            <el-button link type="primary" @click="router.push('/blocked')">查看全部</el-button>
          </div>
          <div v-if="blockedTasks.length" class="blocked-list">
            <div v-for="item in blockedTasks.slice(0, 5)" :key="item.id" class="blocked-item" @click="openTask(item)">
              <div class="blocked-left">
                <el-tag :type="getRiskLevel(calcRiskScore(item)).type" size="small">
                  {{ getRiskLevel(calcRiskScore(item)).label }}
                </el-tag>
                <span class="blocked-title">{{ item.title }}</span>
              </div>
              <div class="blocked-right">
                <span class="text-muted">{{ item.assignee?.nickname || '未分配' }}</span>
                <el-tag type="danger" size="small">阻塞{{ calcBlockedDays(item.blocked_since) }}天</el-tag>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无阻塞任务" :image-size="60" />
        </div>
      </el-col>
      <el-col :span="10">
        <div class="card-box">
          <div class="flex-between mb-16">
            <h3 class="section-title">
              <el-icon color="#4d7bfe"><Folder /></el-icon> 项目概览
            </h3>
            <el-button link type="primary" @click="router.push('/projects')">全部项目</el-button>
          </div>
          <div v-if="projects.length" class="project-list">
            <div v-for="p in projects.slice(0, 5)" :key="p.id" class="project-item" @click="router.push(`/project/${p.id}`)">
              <div class="project-name">{{ p.name }}</div>
              <el-progress
                :percentage="projectProgress(p.id)"
                :stroke-width="8"
                :color="progressColor(projectProgress(p.id))"
              />
            </div>
          </div>
          <el-empty v-else description="暂无项目" :image-size="60" />
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="16">
      <el-col :span="12">
        <div class="card-box">
          <h3 class="section-title mb-16">
            <el-icon color="#4d7bfe"><PieChart /></el-icon> 任务状态分布
          </h3>
          <div ref="statusChartRef" class="chart"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="card-box">
          <h3 class="section-title mb-16">
            <el-icon color="#31c49c"><Histogram /></el-icon> 优先级分布
          </h3>
          <div ref="priorityChartRef" class="chart"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showTaskDetail" :task-id="currentTaskId" @updated="loadData" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getProjects } from '@/api/project'
import { getBlockedTasks } from '@/api/task'
import { getStatistics } from '@/api/report'
import { TASK_STATUS_MAP, PRIORITY_MAP } from '@/constants'
import { calcRiskScore, getRiskLevel, calcBlockedDays } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const router = useRouter()
const loading = ref(false)
const projects = ref([])
const blockedTasks = ref([])
const stats = ref({})
const projectStats = ref({})
const showTaskDetail = ref(false)
const currentTaskId = ref(0)

// 图表
const statusChartRef = ref(null)
const priorityChartRef = ref(null)
let statusChart = null
let priorityChart = null

function renderCharts() {
  if (statusChartRef.value) {
    if (!statusChart) statusChart = echarts.init(statusChartRef.value)
    statusChart.setOption(statusChartOption.value, true)
  }
  if (priorityChartRef.value) {
    if (!priorityChart) priorityChart = echarts.init(priorityChartRef.value)
    priorityChart.setOption(priorityChartOption.value, true)
  }
}

function handleResize() {
  statusChart?.resize()
  priorityChart?.resize()
}

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  statusChart?.dispose()
  priorityChart?.dispose()
  statusChart = null
  priorityChart = null
})

// 数组转map工具：[{status:1,count:5}] => {"1":5}
function arrToMap(arr, keyField) {
  const map = {}
  if (Array.isArray(arr)) {
    arr.forEach((item) => {
      map[item[keyField]] = item.count
    })
  }
  return map
}

// 状态分布饼图
const statusChartOption = computed(() => {
  const byStatus = arrToMap(stats.value.by_status, 'status')
  const data = Object.entries(TASK_STATUS_MAP).map(([key, item]) => ({
    name: item.label,
    value: byStatus[key] || 0,
    itemStyle: { color: item.color },
  })).filter((d) => d.value > 0)
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, icon: 'circle' },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['50%', '45%'],
      label: { show: false },
      data,
    }],
  }
})

// 优先级柱状图
const priorityChartOption = computed(() => {
  const byPriority = arrToMap(stats.value.by_priority, 'priority')
  const data = [
    { name: '高', value: byPriority[1] || 0, color: PRIORITY_MAP[1].color },
    { name: '中', value: byPriority[2] || 0, color: PRIORITY_MAP[2].color },
    { name: '低', value: byPriority[3] || 0, color: PRIORITY_MAP[3].color },
  ]
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: data.map((d) => d.name) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      type: 'bar',
      barWidth: 40,
      data: data.map((d) => ({ value: d.value, itemStyle: { color: d.color } })),
    }],
  }
})

watch([statusChartOption, priorityChartOption], () => {
  nextTick(renderCharts)
})

function projectProgress(projectId) {
  const s = projectStats.value[projectId]
  if (!s) return 0
  return s.completion_rate || 0
}

function progressColor(percentage) {
  if (percentage >= 80) return '#31c49c'
  if (percentage >= 40) return '#f0a020'
  return '#f25555'
}

function openTask(task) {
  currentTaskId.value = task.id
  showTaskDetail.value = true
}

async function loadData() {
  loading.value = true
  try {
    // 加载项目列表
    const projRes = await getProjects()
    projects.value = projRes.data?.list || projRes.data || []

    // 加载阻塞任务
    const blockedRes = await getBlockedTasks({ page: 1, page_size: 10 })
    blockedTasks.value = blockedRes.data?.list || blockedRes.data || []

    // 加载第一个项目的统计（或汇总）
    if (projects.value.length) {
      const firstProject = projects.value[0]
      const statRes = await getStatistics(firstProject.id)
      stats.value = statRes.data || {}
      // 加载各项目统计
      for (const p of projects.value.slice(0, 5)) {
        try {
          const s = await getStatistics(p.id)
          projectStats.value[p.id] = s.data || {}
        } catch { /* 忽略 */ }
      }
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  nextTick(renderCharts)
  loadData()
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.stat-row {
  margin-bottom: 16px;
}

.main-row {
  margin-bottom: 16px;
}

.blocked-list {
  .blocked-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 8px;
    border-bottom: 1px solid $ad-border-light;
    cursor: pointer;
    border-radius: $ad-radius-sm;
    transition: background 0.15s;

    &:hover {
      background: $ad-bg-page;
    }

    .blocked-left {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 1;
      min-width: 0;

      .blocked-title {
        font-size: 13px;
        font-weight: 500;
        color: $ad-text-primary;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .blocked-right {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-shrink: 0;
    }
  }
}

.project-list {
  .project-item {
    padding: 12px 8px;
    border-bottom: 1px solid $ad-border-light;
    cursor: pointer;
    border-radius: $ad-radius-sm;
    transition: background 0.15s;

    &:hover {
      background: $ad-bg-page;
    }

    .project-name {
      font-size: 13px;
      font-weight: 500;
      color: $ad-text-primary;
      margin-bottom: 6px;
    }
  }
}

.chart {
  height: 280px;
}
</style>