<template>
  <div class="project-detail" v-loading="loading">
    <!-- 项目信息头 -->
    <div class="ad-card project-header">
      <div class="project-header-main">
        <div class="project-header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/projects' }">项目列表</el-breadcrumb-item>
            <el-breadcrumb-item>{{ project.name }}</el-breadcrumb-item>
          </el-breadcrumb>
          <h2 class="project-name">{{ project.name }}</h2>
          <p class="project-desc">{{ project.description || '暂无描述' }}</p>
        </div>
        <div class="project-header-right">
          <el-tag :type="statusInfo(project.status).type" effect="light" round>
            {{ statusInfo(project.status).label }}
          </el-tag>
          <span class="project-date">
            <el-icon><Calendar /></el-icon>
            {{ formatDate(project.start_time) }} ~ {{ formatDate(project.end_time) }}
          </span>
        </div>
      </div>

      <!-- 统计卡片 -->
      <div class="project-stats">
        <div class="ad-stat-card stat-item">
          <div class="ad-stat-icon" style="background: #eef2ff; color: #4d7bfe">
            <el-icon :size="24"><List /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value">{{ stats.total_tasks || 0 }}</div>
            <div class="ad-stat-label">总任务</div>
          </div>
        </div>
        <div class="ad-stat-card stat-item">
          <div class="ad-stat-icon" style="background: #e8f8f2; color: #31c49c">
            <el-icon :size="24"><CircleCheck /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #31c49c">{{ stats.completed_tasks || 0 }}</div>
            <div class="ad-stat-label">已完成</div>
          </div>
        </div>
        <div class="ad-stat-card stat-item">
          <div class="ad-stat-icon" style="background: #fef0f0; color: #f25555">
            <el-icon :size="24"><Warning /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #f25555">{{ stats.blocked_tasks || 0 }}</div>
            <div class="ad-stat-label">阻塞中</div>
          </div>
        </div>
        <div class="ad-stat-card stat-item">
          <div class="ad-stat-icon" style="background: #fdf6ec; color: #f0a020">
            <el-icon :size="24"><User /></el-icon>
          </div>
          <div>
            <div class="ad-stat-value" style="color: #f0a020">{{ stats.member_count || 0 }}</div>
            <div class="ad-stat-label">成员数</div>
          </div>
        </div>
      </div>

      <!-- 完成率 -->
      <div class="progress-section" v-if="stats.total_tasks">
        <div class="flex-between mb-8">
          <span class="progress-label">项目完成率</span>
          <span class="progress-value">{{ stats.completion_rate || 0 }}%</span>
        </div>
        <el-progress :percentage="stats.completion_rate || 0" :stroke-width="10" :color="progressColor" :show-text="false" />
      </div>
    </div>

    <!-- 视图模式切换 + 成员 Tab -->
    <div class="ad-card view-switcher">
      <div class="flex-between">
        <el-radio-group v-model="viewMode" size="default">
          <el-radio-button value="tree">
            <el-icon><Menu /></el-icon>&nbsp;任务树
          </el-radio-button>
          <el-radio-button value="kanban">
            <el-icon><Grid /></el-icon>&nbsp;看板
          </el-radio-button>
          <el-radio-button value="gantt">
            <el-icon><DataLine /></el-icon>&nbsp;甘特图
          </el-radio-button>
          <el-radio-button value="members">
            <el-icon><UserFilled /></el-icon>&nbsp;成员
          </el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="showCreateTask = true">
          <el-icon><Plus /></el-icon>&nbsp;新建任务
        </el-button>
      </div>
    </div>

    <!-- ==================== 任务树视图 ==================== -->
    <div v-if="viewMode === 'tree'" class="ad-card">
      <el-table
        v-if="taskTree.length"
        :data="taskTree"
        row-key="id"
        border
        default-expand-all
        :tree-props="{ children: 'children' }"
        size="default"
      >
        <el-table-column prop="title" label="任务名称" min-width="240" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="TASK_STATUS_MAP[row.status]?.type" size="small" effect="light">
              {{ TASK_STATUS_MAP[row.status]?.label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="PRIORITY_MAP[row.priority]?.type" size="small" effect="plain">
              {{ PRIORITY_MAP[row.priority]?.label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="100">
          <template #default="{ row }">
            <div class="assignee-cell">
              <el-avatar :size="20" class="mini-avatar">{{ row.assignee?.nickname?.charAt(0) || '?' }}</el-avatar>
              {{ row.assignee?.nickname || '-' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="进度" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.progress || 0" :stroke-width="6" />
          </template>
        </el-table-column>
        <el-table-column label="截止日期" width="120">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isOverdue(row.due_date, row.status) }">{{ formatDate(row.due_date) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openTaskDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="openAddSub(row)">加子任务</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无任务，点击右上角创建" />
    </div>

    <!-- ==================== 看板视图 ==================== -->
    <div v-if="viewMode === 'kanban'" class="kanban-board">
      <div
        v-for="col in kanbanColumns"
        :key="col.value"
        class="kanban-column"
        :style="{ borderTopColor: col.color }"
        @dragover.prevent
        @drop="handleKanbanDrop($event, col.value)"
      >
        <div class="column-header">
          <span class="column-title" :style="{ color: col.color }">{{ col.label }}</span>
          <el-badge :value="getKanbanTasks(col.value).length" type="info" />
        </div>
        <div class="column-body">
          <div
            v-for="task in getKanbanTasks(col.value)"
            :key="task.id"
            class="task-card"
            :class="{ 'task-card--blocked': task.blocked_reason }"
            draggable="true"
            @dragstart="handleKanbanDragStart($event, task)"
            @click="openTaskDetail(task)"
          >
            <div class="task-card-title">{{ task.title }}</div>
            <div class="task-card-meta">
              <el-tag :type="PRIORITY_MAP[task.priority]?.type" size="small" effect="plain">
                {{ PRIORITY_MAP[task.priority]?.label }}
              </el-tag>
              <el-tag v-if="task.blocked_reason" type="danger" size="small" effect="light">阻塞</el-tag>
              <el-tag v-if="isOverdue(task.due_date, task.status)" type="danger" size="small" effect="plain">逾期</el-tag>
            </div>
            <div class="task-card-footer">
              <span class="task-assignee">{{ task.assignee?.nickname || '未分配' }}</span>
              <span class="task-due" :class="{ 'text-danger': isOverdue(task.due_date, task.status) }">
                {{ formatDate(task.due_date) }}
              </span>
            </div>
          </div>
          <div v-if="!getKanbanTasks(col.value).length" class="empty-column">拖拽任务到此列</div>
        </div>
      </div>
    </div>

    <!-- ==================== 甘特图视图 ==================== -->
    <div v-if="viewMode === 'gantt'" class="ad-card" v-loading="ganttLoading">
      <div class="flex-between mb-16">
        <span class="gantt-tip">任务时间轴视图</span>
        <el-button-group>
          <el-button :type="ganttZoom === 'day' ? 'primary' : ''" size="small" @click="ganttZoom = 'day'">日</el-button>
          <el-button :type="ganttZoom === 'week' ? 'primary' : ''" size="small" @click="ganttZoom = 'week'">周</el-button>
        </el-button-group>
      </div>
      <div v-if="ganttTasks.length" class="gantt-wrap">
        <div class="gantt-left">
          <div class="gantt-left-header">任务名称</div>
          <div v-for="task in ganttTasksWithDepth" :key="task.id" class="gantt-left-row" :style="{ paddingLeft: `${(task.depth || 0) * 16 + 8}px` }">
            <span class="task-name-text" @click="openTaskDetail(task)">
              <el-icon v-if="task.is_milestone" color="#f0a020" size="12"><Star /></el-icon>
              {{ task.title }}
            </span>
          </div>
        </div>
        <div class="gantt-right">
          <div class="gantt-header">
            <div v-for="d in ganttDateHeaders" :key="d.key" class="gantt-header-cell" :class="{ 'is-today': d.isToday }">
              {{ d.label }}
            </div>
          </div>
          <div class="gantt-body">
            <div v-if="ganttTodayOffset >= 0" class="today-line" :style="{ left: `${ganttTodayOffset}px` }"></div>
            <div v-for="d in ganttDateHeaders" :key="'grid-' + d.key" class="grid-line" :style="{ left: `${d.offset}px` }"></div>
            <div v-for="task in ganttTasks" :key="task.id" class="gantt-row">
              <div
                v-if="getGanttBarStyle(task)"
                class="gantt-bar"
                :class="{ 'is-blocked': task.blocked_reason, 'is-overdue': isTaskOverdue(task), 'is-milestone': task.is_milestone }"
                :style="getGanttBarStyle(task)"
                @click="openTaskDetail(task)"
              >
                <div class="gantt-bar-progress" :style="{ width: `${task.progress || 0}%` }"></div>
                <span class="gantt-bar-text">{{ task.progress || 0 }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无甘特图数据" />
    </div>

    <!-- ==================== 成员视图 ==================== -->
    <div v-if="viewMode === 'members'" class="ad-card">
      <div class="flex-between mb-16">
        <span class="section-title">项目成员</span>
        <el-button type="primary" size="small" @click="showAddMember = true">
          <el-icon><Plus /></el-icon>&nbsp;添加成员
        </el-button>
      </div>
      <el-table :data="members" size="default">
        <el-table-column label="成员" min-width="180">
          <template #default="{ row }">
            <div class="member-cell">
              <el-avatar :size="32" class="mini-avatar">{{ row.user?.nickname?.charAt(0) || '?' }}</el-avatar>
              <div>
                <div class="member-name">{{ row.user?.nickname || '-' }}</div>
                <div class="member-phone text-muted">{{ row.user?.phone || '' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 1 ? 'warning' : 'info'" size="small" effect="light">
              {{ row.role === 1 ? '管理员' : '成员' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="加入时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="danger" size="small" @click="handleRemoveMember(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建任务弹窗 -->
    <el-dialog v-model="showCreateTask" :title="parentTask ? `为「${parentTask.title}」添加子任务` : '新建任务'" width="560px">
      <el-form ref="taskFormRef" :model="taskForm" :rules="taskRules" label-width="80px">
        <el-form-item label="任务标题" prop="title">
          <el-input v-model="taskForm.title" placeholder="输入任务标题" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="taskForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-row>
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-select v-model="taskForm.priority" style="width: 100%">
                <el-option v-for="p in PRIORITY_LIST" :key="p.value" :label="p.label" :value="p.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="截止日期">
              <el-date-picker v-model="taskForm.due_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-select v-model="taskForm.assignee_id" clearable filterable placeholder="选择负责人" style="width: 100%">
                <el-option v-for="m in memberOptions" :key="m.id" :label="m.nickname" :value="m.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预估工时">
              <el-input-number v-model="taskForm.estimate" :min="0" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="开始日期">
              <el-date-picker v-model="taskForm.start_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束日期">
              <el-date-picker v-model="taskForm.end_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showCreateTask = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreateTask">创建</el-button>
      </template>
    </el-dialog>

    <!-- 添加成员弹窗 -->
    <el-dialog v-model="showAddMember" title="添加项目成员" width="480px">
      <el-form label-width="80px">
        <el-form-item label="选择用户">
          <el-select v-model="addMemberForm.user_id" filterable placeholder="搜索用户" style="width: 100%">
            <el-option v-for="u in teamMembers" :key="u.id" :label="u.nickname" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="addMemberForm.role" style="width: 100%">
            <el-option label="管理员" :value="1" />
            <el-option label="成员" :value="2" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddMember = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleAddMember">添加</el-button>
      </template>
    </el-dialog>

    <!-- 任务详情弹窗 -->
    <TaskDetailDialog v-model:visible="showTaskDetail" :task-id="currentTaskId" @updated="loadData" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { getProject, getProjectMembers, addProjectMember, removeProjectMember } from '@/api/project'
import { getTaskTree, createTask, getTasks, updateTaskStatus } from '@/api/task'
import { getStatistics, getGanttChart } from '@/api/report'
import { getTeam } from '@/api/team'
import { TASK_STATUS_MAP, PRIORITY_MAP, PRIORITY_LIST, PROJECT_STATUS_MAP } from '@/constants'
import { formatDate, formatDateTime, isOverdue } from '@/utils'
import TaskDetailDialog from '@/components/TaskDetailDialog.vue'

const route = useRoute()
const projectId = computed(() => Number(route.params.id))

const loading = ref(false)
const saving = ref(false)
const viewMode = ref('tree') // tree | kanban | gantt | members
const project = ref({})
const stats = ref({})
const taskTree = ref([])
const members = ref([])
const teamMembers = ref([])
const showCreateTask = ref(false)
const showAddMember = ref(false)
const showTaskDetail = ref(false)
const currentTaskId = ref(0)
const parentTask = ref(null)
const taskFormRef = ref()

// 看板
const kanbanTasks = ref([])
const dragTask = ref(null)
const kanbanColumns = Object.entries(TASK_STATUS_MAP).map(([key, item]) => ({
  value: Number(key),
  label: item.label,
  color: item.color,
}))

// 甘特图
const ganttLoading = ref(false)
const ganttTasks = ref([])
const ganttZoom = ref('week')
const GANTT_CELL_WIDTH = 40

const taskForm = reactive({
  title: '', description: '', priority: 2, assignee_id: 0,
  due_date: '', estimate: 0, start_date: '', end_date: '',
})
const addMemberForm = reactive({ user_id: 0, role: 2 })
const taskRules = {
  title: [{ required: true, message: '请输入任务标题', trigger: 'blur' }],
}

const memberOptions = computed(() =>
  members.value.map((m) => ({ id: m.user_id, nickname: m.user?.nickname || '-' }))
)

const progressColor = computed(() => {
  const rate = stats.value.completion_rate || 0
  if (rate >= 80) return '#31c49c'
  if (rate >= 40) return '#f0a020'
  return '#f25555'
})

function statusInfo(status) {
  return PROJECT_STATUS_MAP[status] || { label: '-', type: 'info' }
}

function openTaskDetail(row) {
  currentTaskId.value = row.id
  showTaskDetail.value = true
}

function openAddSub(row) {
  parentTask.value = row
  resetTaskForm()
  showCreateTask.value = true
}

function resetTaskForm() {
  taskForm.title = ''
  taskForm.description = ''
  taskForm.priority = 2
  taskForm.assignee_id = 0
  taskForm.due_date = ''
  taskForm.estimate = 0
  taskForm.start_date = ''
  taskForm.end_date = ''
}

// ==================== 看板逻辑 ====================
function getKanbanTasks(status) {
  return kanbanTasks.value.filter((t) => t.status === status)
}

function handleKanbanDragStart(event, task) {
  dragTask.value = task
  event.dataTransfer.effectAllowed = 'move'
}

async function handleKanbanDrop(event, newStatus) {
  const task = dragTask.value
  if (!task || task.status === newStatus) return
  const oldStatus = task.status
  try {
    task.status = newStatus
    await updateTaskStatus(task.id, { status: newStatus })
    ElMessage.success(`已移动到「${TASK_STATUS_MAP[newStatus]?.label}」`)
  } catch {
    task.status = oldStatus
  } finally {
    dragTask.value = null
  }
}

async function loadKanbanTasks() {
  try {
    const res = await getTasks({ project_id: projectId.value, page: 1, page_size: 200 })
    kanbanTasks.value = res.data?.list || []
  } catch { /* 忽略 */ }
}

// ==================== 甘特图逻辑 ====================
const ganttTasksWithDepth = computed(() => {
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

const ganttDateRange = computed(() => {
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
  minDate = minDate.subtract(3, 'day')
  maxDate = maxDate.add(3, 'day')
  return { start: minDate, end: maxDate, days: maxDate.diff(minDate, 'day') + 1 }
})

const ganttDateHeaders = computed(() => {
  if (!ganttDateRange.value) return []
  const { start, days } = ganttDateRange.value
  const headers = []
  const step = ganttZoom.value === 'day' ? 1 : 7
  for (let i = 0; i < days; i += step) {
    const d = start.add(i, 'day')
    headers.push({
      key: d.format('YYYY-MM-DD'),
      label: step === 1 ? d.format('MM/DD') : `${d.format('MM/DD')}~${d.add(6, 'day').format('MM/DD')}`,
      offset: i * GANTT_CELL_WIDTH,
      isToday: d.isSame(dayjs(), 'day'),
    })
  }
  return headers
})

const ganttTodayOffset = computed(() => {
  if (!ganttDateRange.value) return -1
  const diff = dayjs().startOf('day').diff(ganttDateRange.value.start.startOf('day'), 'day')
  return diff >= 0 && diff <= ganttDateRange.value.days ? diff * GANTT_CELL_WIDTH : -1
})

function getGanttBarStyle(task) {
  if (!ganttDateRange.value || !task.start_date || !task.end_date) return null
  const startOffset = dayjs(task.start_date).startOf('day').diff(ganttDateRange.value.start.startOf('day'), 'day')
  const endOffset = dayjs(task.end_date).startOf('day').diff(ganttDateRange.value.start.startOf('day'), 'day')
  if (startOffset < 0 || endOffset < startOffset) return null
  const width = Math.max((endOffset - startOffset + 1) * GANTT_CELL_WIDTH - 4, 20)
  return { left: `${startOffset * GANTT_CELL_WIDTH + 2}px`, width: `${width}px` }
}

function isTaskOverdue(task) {
  return isOverdue(task.end_date, task.status)
}

async function loadGanttTasks() {
  ganttLoading.value = true
  try {
    const res = await getGanttChart(projectId.value)
    ganttTasks.value = res.data || []
  } finally {
    ganttLoading.value = false
  }
}

// 视图切换时加载对应数据
watch(viewMode, (mode) => {
  if (mode === 'kanban' && !kanbanTasks.value.length) loadKanbanTasks()
  if (mode === 'gantt') loadGanttTasks()
})

// ==================== CRUD 操作 ====================
async function handleCreateTask() {
  await taskFormRef.value.validate()
  saving.value = true
  try {
    await createTask({
      project_id: projectId.value,
      parent_id: parentTask.value?.id || 0,
      title: taskForm.title,
      description: taskForm.description,
      priority: taskForm.priority,
      assignee_id: taskForm.assignee_id || 0,
      due_date: taskForm.due_date || null,
      start_date: taskForm.start_date || null,
      end_date: taskForm.end_date || null,
      estimate: taskForm.estimate,
    })
    ElMessage.success('任务创建成功')
    showCreateTask.value = false
    parentTask.value = null
    await loadData()
    if (viewMode.value === 'kanban') loadKanbanTasks()
    if (viewMode.value === 'gantt') loadGanttTasks()
  } finally {
    saving.value = false
  }
}

async function handleAddMember() {
  if (!addMemberForm.user_id) {
    ElMessage.warning('请选择用户')
    return
  }
  saving.value = true
  try {
    await addProjectMember(projectId.value, { user_id: addMemberForm.user_id, role: addMemberForm.role })
    ElMessage.success('添加成功')
    showAddMember.value = false
    addMemberForm.user_id = 0
    addMemberForm.role = 2
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleRemoveMember(row) {
  await ElMessageBox.confirm('确定移除该成员吗？', '提示', { type: 'warning' })
  await removeProjectMember(projectId.value, row.user_id)
  ElMessage.success('已移除')
  await loadData()
}

async function loadData() {
  loading.value = true
  try {
    const [projRes, statRes, treeRes, memberRes] = await Promise.all([
      getProject(projectId.value),
      getStatistics(projectId.value),
      getTaskTree(projectId.value),
      getProjectMembers(projectId.value),
    ])
    project.value = projRes.data || {}
    stats.value = statRes.data || {}
    taskTree.value = treeRes.data || []
    members.value = memberRes.data || []

    try {
      const teamRes = await getTeam()
      teamMembers.value = teamRes.data?.members || []
    } catch { /* 忽略 */ }
  } finally {
    loading.value = false
  }
}

watch(showCreateTask, (val) => {
  if (val && !parentTask.value) resetTaskForm()
})

onMounted(loadData)
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.project-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

// 项目头部
.project-header {
  .project-header-main {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;
  }

  .project-name {
    font-size: 24px;
    font-weight: 800;
    color: $ad-text-primary;
    margin: 8px 0 4px;
    letter-spacing: -0.02em;
  }

  .project-desc {
    font-size: 13px;
    color: $ad-text-secondary;
  }

  .project-header-right {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 8px;
    flex-shrink: 0;
  }

  .project-date {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: $ad-text-secondary;
  }
}

.project-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;

  .stat-item {
    padding: 16px 20px;
    margin-bottom: 0;
  }
}

.progress-section {
  padding-top: 12px;
  border-top: 1px solid $ad-border-light;

  .progress-label {
    font-size: 14px;
    font-weight: 500;
    color: $ad-text-regular;
  }

  .progress-value {
    font-size: 16px;
    font-weight: 700;
    color: $ad-primary;
  }
}

// 视图切换
.view-switcher {
  padding: 16px 24px;
}

// 成员/负责人单元格
.assignee-cell, .member-cell {
  display: flex;
  align-items: center;
  gap: 8px;

  .mini-avatar {
    background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
    color: #fff;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }
}

.member-cell {
  .member-name {
    font-size: 14px;
    font-weight: 500;
    color: $ad-text-primary;
  }
  .member-phone {
    font-size: 12px;
  }
}

// ==================== 看板 ====================
.kanban-board {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  min-height: 500px;
}

.kanban-column {
  flex: 1;
  min-width: 230px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: $ad-radius-lg;
  border-top: 3px solid $ad-info;
  display: flex;
  flex-direction: column;
  box-shadow: $ad-shadow-sm;
  border: 1px solid rgba(0, 0, 0, 0.02);
  border-top: 3px solid;

  .column-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 10px;

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
  border-radius: $ad-radius;
  padding: 12px 14px;
  margin-bottom: 8px;
  cursor: grab;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  transition: all 0.2s;
  border: 1px solid rgba(0, 0, 0, 0.03);

  &:hover {
    box-shadow: $ad-shadow-hover;
    transform: translateY(-1px);
  }

  &:active {
    cursor: grabbing;
  }

  &--blocked {
    border-left: 3px solid $ad-danger;
  }

  .task-card-title {
    font-size: 13px;
    font-weight: 500;
    margin-bottom: 8px;
    line-height: 1.5;
    color: $ad-text-primary;
  }

  .task-card-meta {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
    margin-bottom: 8px;
  }

  .task-card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: $ad-text-secondary;

    .task-due {
      font-size: 12px;
    }
  }
}

.empty-column {
  text-align: center;
  color: $ad-text-placeholder;
  font-size: 12px;
  padding: 20px 0;
}

// ==================== 甘特图 ====================
.gantt-tip {
  font-size: 13px;
  color: $ad-text-secondary;
}

.gantt-wrap {
  display: flex;
  overflow-x: auto;
  border: 1px solid $ad-border-light;
  border-radius: $ad-radius;
}

.gantt-left {
  flex-shrink: 0;
  width: 240px;
  border-right: 2px solid $ad-border;

  .gantt-left-header {
    height: 36px;
    display: flex;
    align-items: center;
    padding: 0 12px;
    font-size: 13px;
    font-weight: 600;
    background: $ad-bg-page;
    border-bottom: 1px solid $ad-border-light;
    color: $ad-text-primary;
  }

  .gantt-left-row {
    height: 36px;
    display: flex;
    align-items: center;
    font-size: 12px;
    border-bottom: 1px solid $ad-border-light;
    overflow: hidden;

    .task-name-text {
      cursor: pointer;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      display: flex;
      align-items: center;
      gap: 4px;
      color: $ad-text-regular;

      &:hover {
        color: $ad-primary;
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
    background: $ad-bg-page;
    border-bottom: 1px solid $ad-border-light;

    .gantt-header-cell {
      flex-shrink: 0;
      width: 40px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 11px;
      color: $ad-text-secondary;
      border-right: 1px solid $ad-border-light;
      white-space: nowrap;
      overflow: hidden;

      &.is-today {
        background: $ad-primary-lighter;
        color: $ad-primary;
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
      background: $ad-border-light;
    }

    .today-line {
      position: absolute;
      top: 0;
      bottom: 0;
      width: 2px;
      background: $ad-danger;
      z-index: 2;
    }

    .gantt-row {
      height: 36px;
      position: relative;
      border-bottom: 1px solid $ad-border-light;

      .gantt-bar {
        position: absolute;
        top: 8px;
        height: 20px;
        border-radius: 4px;
        background: $ad-primary;
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
          background: $ad-danger;
        }

        &.is-overdue {
          box-shadow: 0 0 0 2px rgba(242, 85, 85, 0.4);
        }

        &.is-milestone {
          background: $ad-warning;
          border-radius: 50%;
          width: 20px !important;
          height: 20px;
        }
      }
    }
  }
}
</style>