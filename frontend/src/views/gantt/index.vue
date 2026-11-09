<template>
  <div class="gantt-page">
    <!-- 工具栏 -->
    <div class="card-box mb-16">
      <div class="flex-between">
        <h3 class="page-title">甘特图</h3>
        <div class="toolbar">
          <el-select v-model="projectId" placeholder="选择项目" style="width: 200px" @change="loadData">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-button-group>
            <el-button :type="zoomLevel === 'day' ? 'primary' : ''" size="small" @click="setZoom('day')">日</el-button>
            <el-button :type="zoomLevel === 'week' ? 'primary' : ''" size="small" @click="setZoom('week')">周</el-button>
          </el-button-group>
        </div>
      </div>
    </div>

    <!-- 甘特图 -->
    <div class="card-box" v-loading="loading">
      <div v-if="ganttTasks.length" class="gantt-wrap">
        <!-- 左侧任务名列表 -->
        <div class="gantt-left">
          <div class="gantt-left-header">任务名称</div>
          <div v-for="task in tasksWithDepth" :key="task.id" class="gantt-left-row" :style="{ paddingLeft: `${(task.depth || 0) * 16 + 8}px` }">
            <span class="task-name-text" @click="openDetail(task)">
              <el-icon v-if="task.is_milestone" color="#f0a020" size="12"><Star /></el-icon>
              {{ task.title }}
            </span>
          </div>
        </div>

        <!-- 右侧时间轴 -->
        <div class="gantt-right" ref="ganttRightRef">
          <!-- 日期头部 -->
          <div class="gantt-header">
            <div
              v-for="d in dateHeaders"
              :key="d.key"
              class="gantt-header-cell"
              :class="{ 'is-today': d.isToday }"
            >
              {{ d.label }}
            </div>
          </div>
          <!-- 任务条 -->
          <div class="gantt-body">
            <!-- 今日竖线 -->
            <div v-if="todayOffset >= 0" class="today-line" :style="{ left: `${todayOffset}px` }"></div>
            <!-- 网格竖线 -->
            <div
              v-for="d in dateHeaders"
              :key="'grid-' + d.key"
              class="grid-line"
              :style="{ left: `${d.offset}px` }"
            ></div>
            <!-- 任务行 -->
            <div v-for="task in ganttTasks" :key="task.id" class="gantt-row">
              <div
                v-if="getBarStyle(task)"
                class="gantt-bar"
                :class="{ 'is-blocked': task.blocked_reason, 'is-overdue': isTaskOverdue(task), 'is-milestone': task.is_milestone }"
                :style="getBarStyle(task)"
                @click="openDetail(task)"
              >
                <div class="gantt-bar-progress" :style="{ width: `${task.progress || 0}%` }"></div>
                <span class="gantt-bar-text">{{ task.progress || 0 }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <el-empty v-else description="请选择项目查看甘特图" />
    </div>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showDetail" :task-id="currentTaskId" @updated="loadData" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import dayjs from 'dayjs'
import { getGanttChart } from '@/api/report'
import { getProjects } from '@/api/project'
import { TASK_STATUS } from '@/constants'
import { isOverdue } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const loading = ref(false)
const projectId = ref(null)
const projects = ref([])
const ganttTasks = ref([])
const showDetail = ref(false)
const currentTaskId = ref(0)
const zoomLevel = ref('week')
const ganttRightRef = ref()

const CELL_WIDTH = 40

// 计算任务深度
const tasksWithDepth = computed(() => {
  const map = {}
  ganttTasks.value.forEach((t) => { map[t.id] = t })
  return ganttTasks.value.map((t) => {
    let depth = 0
    let p = t
    while (p.parent_id && map[p.parent_id]) {
      depth++
      p = map[p.parent_id]
    }
    return { ...t, depth }
  })
})

// 计算日期范围
const dateRange = computed(() => {
  if (!ganttTasks.value.length) return null
  let minDate = null
  let maxDate = null
  ganttTasks.value.forEach((t) => {
    if (t.start_date) {
      const d = dayjs(t.start_date)
      if (!minDate || d.isBefore(minDate)) minDate = d
    }
    if (t.end_date) {
      const d = dayjs(t.end_date)
      if (!maxDate || d.isAfter(maxDate)) maxDate = d
    }
  })
  if (!minDate) minDate = dayjs().subtract(7, 'day')
  if (!maxDate) maxDate = dayjs().add(30, 'day')
  // 扩展边距
  minDate = minDate.subtract(3, 'day')
  maxDate = maxDate.add(3, 'day')
  return { start: minDate, end: maxDate, days: maxDate.diff(minDate, 'day') + 1 }
})

// 日期头部
const dateHeaders = computed(() => {
  if (!dateRange.value) return []
  const { start, days } = dateRange.value
  const headers = []
  const step = zoomLevel.value === 'day' ? 1 : 7
  for (let i = 0; i < days; i += step) {
    const d = start.add(i, 'day')
    const isToday = d.isSame(dayjs(), 'day')
    headers.push({
      key: d.format('YYYY-MM-DD'),
      label: step === 1 ? d.format('MM/DD') : `${d.format('MM/DD')}~${d.add(6, 'day').format('MM/DD')}`,
      offset: i * CELL_WIDTH,
      isToday,
    })
  }
  return headers
})

// 今日竖线位置
const todayOffset = computed(() => {
  if (!dateRange.value) return -1
  const diff = dayjs().startOf('day').diff(dateRange.value.start.startOf('day'), 'day')
  return diff >= 0 && diff <= dateRange.value.days ? diff * CELL_WIDTH : -1
})

function setZoom(level) {
  zoomLevel.value = level
}

function getBarStyle(task) {
  if (!dateRange.value || !task.start_date || !task.end_date) return null
  const startOffset = dayjs(task.start_date).startOf('day').diff(dateRange.value.start.startOf('day'), 'day')
  const endOffset = dayjs(task.end_date).startOf('day').diff(dateRange.value.start.startOf('day'), 'day')
  if (startOffset < 0 || endOffset < startOffset) return null
  const width = Math.max((endOffset - startOffset + 1) * CELL_WIDTH - 4, 20)
  return {
    left: `${startOffset * CELL_WIDTH + 2}px`,
    width: `${width}px`,
  }
}

function isTaskOverdue(task) {
  return isOverdue(task.end_date, task.status)
}

function openDetail(task) {
  currentTaskId.value = task.id
  showDetail.value = true
}

async function loadData() {
  if (!projectId.value) {
    ganttTasks.value = []
    return
  }
  loading.value = true
  try {
    const res = await getGanttChart(projectId.value)
    ganttTasks.value = res.data || []
  } finally {
    loading.value = false
  }
}

async function loadProjects() {
  try {
    const res = await getProjects()
    projects.value = res.data?.list || res.data || []
    // 自动选择第一个项目
    if (projects.value.length && !projectId.value) {
      projectId.value = projects.value[0].id
      await loadData()
    }
  } catch { /* 忽略 */ }
}

onMounted(loadProjects)
</script>

<style lang="scss" scoped>
.page-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
}

.gantt-wrap {
  display: flex;
  overflow-x: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.gantt-left {
  flex-shrink: 0;
  width: 240px;
  border-right: 2px solid #dcdfe6;

  .gantt-left-header {
    height: 36px;
    display: flex;
    align-items: center;
    padding: 0 12px;
    font-size: 13px;
    font-weight: 600;
    background: #f5f7fa;
    border-bottom: 1px solid #ebeef5;
  }

  .gantt-left-row {
    height: 36px;
    display: flex;
    align-items: center;
    font-size: 12px;
    border-bottom: 1px solid #f0f0f0;
    overflow: hidden;

    .task-name-text {
      cursor: pointer;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      display: flex;
      align-items: center;
      gap: 4px;

      &:hover {
        color: #4d7bfe;
      }
    }
  }
}

.gantt-right {
  flex: 1;
  position: relative;
  overflow-x: auto;

  .gantt-header {
    display: flex;
    height: 36px;
    background: #f5f7fa;
    border-bottom: 1px solid #ebeef5;

    .gantt-header-cell {
      flex-shrink: 0;
      width: 40px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 11px;
      color: #606266;
      border-right: 1px solid #ebeef5;
      white-space: nowrap;
      overflow: hidden;

      &.is-today {
        background: #ecf5ff;
        color: #4d7bfe;
        font-weight: 600;
      }
    }
  }

  .gantt-body {
    position: relative;

    .grid-line {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 1px;
      background: #f0f0f0;
    }

    .today-line {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 2px;
      background: #f25555;
      z-index: 2;
    }

    .gantt-row {
      height: 36px;
      position: relative;
      border-bottom: 1px solid #f0f0f0;

      .gantt-bar {
        position: absolute;
        top: 8px;
        height: 20px;
        border-radius: 4px;
        background: #4d7bfe;
        cursor: pointer;
        overflow: hidden;
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1;

        .gantt-bar-progress {
          position: absolute;
          left: 0;
          top: 0;
          height: 100%;
          background: rgba(255, 255, 255, 0.35);
        }

        .gantt-bar-text {
          position: relative;
          font-size: 11px;
          color: #fff;
          white-space: nowrap;
        }

        &.is-blocked {
          background: #f25555;
        }

        &.is-overdue {
          box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.4);
        }

        &.is-milestone {
          background: #f0a020;
          border-radius: 50%;
          width: 20px !important;
          height: 20px;
        }
      }
    }
  }
}
</style>