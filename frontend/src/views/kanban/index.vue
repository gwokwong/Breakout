<template>
  <div class="kanban-page">
    <!-- 顶部工具栏 -->
    <div class="card-box mb-16">
      <div class="flex-between">
        <h3 class="page-title">看板</h3>
        <div class="toolbar">
          <el-select v-model="projectId" clearable placeholder="全部项目" style="width: 180px" @change="loadTasks">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-button @click="loadTasks">
            <el-icon><Refresh /></el-icon>&nbsp;刷新
          </el-button>
        </div>
      </div>
    </div>

    <!-- 看板列 -->
    <div class="kanban-board" v-loading="loading">
      <div
        v-for="col in columns"
        :key="col.value"
        class="kanban-column"
        :style="{ borderTopColor: col.color }"
        @dragover.prevent
        @drop="handleDrop($event, col.value)"
      >
        <div class="column-header">
          <span class="column-title" :style="{ color: col.color }">{{ col.label }}</span>
          <el-badge :value="getColumnTasks(col.value).length" type="info" />
        </div>
        <div class="column-body">
          <div
            v-for="task in getColumnTasks(col.value)"
            :key="task.id"
            class="task-card"
            :class="{ 'task-card--blocked': task.blocked_reason }"
            draggable="true"
            @dragstart="handleDragStart($event, task)"
            @click="openDetail(task)"
          >
            <div class="task-card-title">{{ task.title }}</div>
            <div class="task-card-meta">
              <el-tag :type="PRIORITY_MAP[task.priority]?.type" size="small">{{ PRIORITY_MAP[task.priority]?.label }}</el-tag>
              <el-tag v-if="task.blocked_reason" type="danger" size="small">阻塞</el-tag>
              <el-tag v-if="isOverdue(task.due_date, task.status)" type="danger" size="small" effect="plain">逾期</el-tag>
            </div>
            <div class="task-card-footer">
              <span class="task-assignee">{{ task.assignee?.nickname || '未分配' }}</span>
              <span class="task-due" :class="{ 'text-danger': isOverdue(task.due_date, task.status) }">
                {{ formatDate(task.due_date) }}
              </span>
            </div>
          </div>
          <div v-if="!getColumnTasks(col.value).length" class="empty-column">
            拖拽任务到此列
          </div>
        </div>
      </div>
    </div>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showDetail" :task-id="currentTaskId" @updated="loadTasks" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getTasks, updateTaskStatus } from '@/api/task'
import { getProjects } from '@/api/project'
import { TASK_STATUS_MAP, PRIORITY_MAP, canTransition } from '@/constants'
import { formatDate, isOverdue } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const loading = ref(false)
const projectId = ref(null)
const projects = ref([])
const tasks = ref([])
const showDetail = ref(false)
const currentTaskId = ref(0)
const dragTask = ref(null)

// 6列看板
const columns = Object.entries(TASK_STATUS_MAP).map(([key, item]) => ({
  value: Number(key),
  label: item.label,
  color: item.color,
}))

function getColumnTasks(status) {
  return tasks.value.filter((t) => t.status === status)
}

function handleDragStart(event, task) {
  dragTask.value = task
  event.dataTransfer.effectAllowed = 'move'
}

async function handleDrop(event, newStatus) {
  const task = dragTask.value
  if (!task || task.status === newStatus) return
  const oldStatus = task.status
  // 前端状态流转校验，与后端规则一致
  if (!canTransition(oldStatus, newStatus, task.reviewer_id || 0)) {
    ElMessage.warning(`不允许从「${TASK_STATUS_MAP[oldStatus]?.label}」移动到「${TASK_STATUS_MAP[newStatus]?.label}」`)
    dragTask.value = null
    return
  }
  try {
    // 乐观更新
    task.status = newStatus
    await updateTaskStatus(task.id, { status: newStatus })
    ElMessage.success(`已移动到「${TASK_STATUS_MAP[newStatus]?.label}」`)
  } catch {
    // 回滚
    task.status = oldStatus
  } finally {
    dragTask.value = null
  }
}

function openDetail(task) {
  currentTaskId.value = task.id
  showDetail.value = true
}

async function loadTasks() {
  loading.value = true
  try {
    // 加载较多任务以覆盖各状态
    const params = { page: 1, page_size: 200 }
    if (projectId.value) params.project_id = projectId.value
    const res = await getTasks(params)
    tasks.value = res.data?.list || []
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
  loadTasks()
})
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
}

.kanban-board {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  min-height: 500px;
}

.kanban-column {
  flex: 1;
  min-width: 240px;
  background: #f5f7fa;
  border-radius: 8px;
  border-top: 3px solid #909399;
  display: flex;
  flex-direction: column;

  .column-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px 8px;

    .column-title {
      font-size: 14px;
      font-weight: 600;
    }
  }

  .column-body {
    flex: 1;
    padding: 4px 10px 12px;
    overflow-y: auto;
  }
}

.task-card {
  background: #fff;
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 8px;
  cursor: grab;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }

  &:active {
    cursor: grabbing;
  }

  &--blocked {
    border-left: 3px solid #f25555;
  }

  .task-card-title {
    font-size: 13px;
    font-weight: 500;
    margin-bottom: 6px;
    line-height: 1.4;
  }

  .task-card-meta {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
    margin-bottom: 6px;
  }

  .task-card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: #909399;

    .task-due {
      font-size: 12px;
    }
  }
}

.empty-column {
  text-align: center;
  color: #c0c4cc;
  font-size: 12px;
  padding: 20px 0;
}

.text-danger {
  color: #f25555;
}
</style>