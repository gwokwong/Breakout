<template>
  <el-dialog v-model="visible" title="任务详情" width="800px" top="5vh" @open="loadData">
    <div v-loading="loading">
      <!-- 基本信息 -->
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="任务标题" :span="2">
          <div class="title-row">
            <span v-if="!editing" class="task-title">{{ task.title }}</span>
            <el-input v-else v-model="editForm.title" placeholder="任务标题" />
            <el-button v-if="!editing" link type="primary" @click="startEdit">编辑</el-button>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusInfo.type" size="small">{{ statusInfo.label }}</el-tag>
          <el-tag v-if="task.blocked_reason" type="danger" size="small" style="margin-left: 4px">阻塞中</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="priorityInfo.type" size="small">{{ priorityInfo.label }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="负责人">
          {{ task.assignee?.nickname || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建人">
          {{ task.creator?.nickname || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="复核人">
          {{ task.reviewer?.nickname || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="验收人">
          {{ task.acceptor?.nickname || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="截止日期">
          <span :class="{ 'text-danger': overdue }">{{ formatDate(task.due_date) }}</span>
          <el-tag v-if="overdue" type="danger" size="small" style="margin-left: 4px">逾期</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="预估工时">{{ task.estimate || 0 }}h / 实际: {{ task.actual || 0 }}h</el-descriptions-item>
        <el-descriptions-item label="里程碑">
          <el-tag v-if="task.is_milestone" type="warning" size="small">里程碑</el-tag>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(task.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          <div v-if="!editing" class="task-desc">{{ task.description || '暂无描述' }}</div>
          <el-input v-else v-model="editForm.description" type="textarea" :rows="3" />
        </el-descriptions-item>
        <el-descriptions-item v-if="task.blocked_reason" label="阻塞原因" :span="2">
          <div class="blocked-reason">
            <el-icon color="#f25555"><Warning /></el-icon>
            {{ task.blocked_reason }}
            <span class="text-muted">（阻塞{{ blockedDays }}天）</span>
          </div>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 编辑模式下的表单 -->
      <div v-if="editing" class="edit-form">
        <el-form :model="editForm" label-width="80px" size="small">
          <el-row :gutter="12">
            <el-col :span="8">
              <el-form-item label="优先级">
                <el-select v-model="editForm.priority" style="width: 100%">
                  <el-option v-for="p in PRIORITY_LIST" :key="p.value" :label="p.label" :value="p.value" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="负责人">
                <el-select v-model="editForm.assignee_id" style="width: 100%" clearable filterable>
                  <el-option v-for="m in memberList" :key="m.id" :label="m.nickname" :value="m.id" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="截止日期">
                <el-date-picker v-model="editForm.due_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="复核人">
                <el-select v-model="editForm.reviewer_id" style="width: 100%" clearable filterable>
                  <el-option v-for="m in memberList" :key="m.id" :label="m.nickname" :value="m.id" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="验收人">
                <el-select v-model="editForm.acceptor_id" style="width: 100%" clearable filterable>
                  <el-option v-for="m in memberList" :key="m.id" :label="m.nickname" :value="m.id" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="预估工时">
                <el-input-number v-model="editForm.estimate" :min="0" :step="0.5" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <div class="edit-actions">
            <el-button type="primary" size="small" :loading="saving" @click="saveEdit">保存</el-button>
            <el-button size="small" @click="editing = false">取消</el-button>
          </div>
        </el-form>
      </div>

      <!-- 操作按钮区 -->
      <div class="action-bar">
        <el-button v-if="task.status === 1" type="primary" size="small" @click="changeStatus(2)">开始任务</el-button>
        <el-button v-if="task.status === 2" type="warning" size="small" @click="changeStatus(3)">提交复核</el-button>
        <el-button v-if="task.status === 2" type="success" size="small" @click="changeStatus(4)">提交验收</el-button>
        <!-- 复核操作 -->
        <el-button v-if="task.status === 3" type="success" size="small" @click="handleReview(true)">复核通过</el-button>
        <el-button v-if="task.status === 3" type="danger" size="small" @click="handleReview(false)">复核不通过</el-button>
        <!-- 验收操作 -->
        <el-button v-if="task.status === 4" type="success" size="small" @click="handleAccept(true)">验收通过</el-button>
        <el-button v-if="task.status === 4" type="danger" size="small" @click="handleAccept(false)">验收不通过</el-button>
        <!-- 关闭 -->
        <el-button v-if="task.status === 5" type="info" size="small" @click="showCloseDialog = true">关闭任务</el-button>
        <!-- 阻塞/解除 -->
        <el-button v-if="!task.blocked_reason && task.status < 5" type="danger" plain size="small" @click="showBlockDialog = true">
          标记阻塞
        </el-button>
        <el-button v-if="task.blocked_reason" type="success" plain size="small" @click="handleUnblock">解除阻塞</el-button>
        <!-- AI操作 -->
        <el-dropdown @command="handleAiCommand" style="margin-left: 8px">
          <el-button type="primary" plain size="small">
            AI助手<el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="decompose">AI拆解子任务</el-dropdown-item>
              <el-dropdown-item command="generateDesc">AI生成描述</el-dropdown-item>
              <el-dropdown-item command="suggestAssignee">AI推荐负责人</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <!-- Tab区域：评论 / 附件 -->
      <el-tabs v-model="activeTab" style="margin-top: 16px">
        <el-tab-pane :label="`评论 (${comments.length})`" name="comments">
          <div class="comment-list">
            <div v-for="c in comments" :key="c.id" class="comment-item">
              <el-avatar :size="28">{{ c.user?.nickname?.charAt(0) || '?' }}</el-avatar>
              <div class="comment-body">
                <div class="comment-header">
                  <span class="comment-author">{{ c.user?.nickname || '未知' }}</span>
                  <span class="comment-time">{{ formatDateTime(c.created_at) }}</span>
                </div>
                <div class="comment-content">{{ c.content }}</div>
              </div>
            </div>
            <el-empty v-if="!comments.length" description="暂无评论" :image-size="60" />
          </div>
          <div class="comment-input">
            <el-input v-model="commentText" placeholder="输入评论..." />
            <el-button type="primary" :loading="commenting" @click="submitComment">发送</el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`附件 (${attachments.length})`" name="attachments">
          <el-upload
            :show-file-list="false"
            :http-request="handleUpload"
            multiple
            style="margin-bottom: 12px"
          >
            <el-button type="primary" plain size="small">
              <el-icon><Upload /></el-icon>&nbsp;上传附件
            </el-button>
          </el-upload>
          <el-table :data="attachments" size="small" border>
            <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
            <el-table-column label="大小" width="90">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column prop="file_type" label="类型" width="80" />
            <el-table-column label="上传时间" width="140">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="handleDownload(row)">下载</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteAttachment(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 关闭任务弹窗 -->
    <el-dialog v-model="showCloseDialog" title="关闭任务" width="480px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="关闭证据" required>
          <el-input v-model="closeEvidence" type="textarea" :rows="3" placeholder="请填写关闭证据（如交付物链接、验证结果等）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCloseDialog = false">取消</el-button>
        <el-button type="primary" :loading="operating" @click="handleCloseTask">确认关闭</el-button>
      </template>
    </el-dialog>

    <!-- 标记阻塞弹窗 -->
    <el-dialog v-model="showBlockDialog" title="标记阻塞" width="480px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="阻塞原因" required>
          <el-input v-model="blockReason" type="textarea" :rows="3" placeholder="请描述阻塞原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBlockDialog = false">取消</el-button>
        <el-button type="danger" :loading="operating" @click="handleBlock">确认阻塞</el-button>
      </template>
    </el-dialog>

    <!-- 复核/验收意见弹窗 -->
    <el-dialog v-model="showReviewDialog" :title="reviewTitle" width="480px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="意见">
          <el-input v-model="reviewComment" type="textarea" :rows="3" placeholder="填写意见（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showReviewDialog = false">取消</el-button>
        <el-button :type="reviewPass ? 'success' : 'danger'" :loading="operating" @click="submitReview">
          确认{{ reviewPass ? '通过' : '不通过' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- AI拆解弹窗 -->
    <AiDecomposeDialog
      v-model:visible="showAiDecompose"
      :task-id="task.id"
      :project-id="task.project_id"
      :parent-id="task.id"
      :default-title="task.title"
      :default-desc="task.description"
      @created="loadData"
    />
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { TASK_STATUS_MAP, PRIORITY_MAP, PRIORITY_LIST } from '@/constants'
import { formatDate, formatDateTime, formatFileSize, calcBlockedDays, isOverdue, downloadBlob } from '@/utils'
import {
  getTask, updateTask, updateTaskStatus, markBlocked, unblockTask,
  reviewTask, acceptTask, closeTask,
} from '@/api/task'
import { getComments, addComment } from '@/api/comment'
import { getAttachments, uploadAttachment, downloadAttachment, deleteAttachment } from '@/api/attachment'
import { generateTaskDesc, suggestAssignee } from '@/api/ai'
import { getProjectMembers } from '@/api/project'
import AiDecomposeDialog from './AiDecomposeDialog.vue'

const props = defineProps({
  taskId: { type: Number, default: 0 },
})
const emit = defineEmits(['updated'])

const visible = defineModel('visible', { type: Boolean, default: false })

const loading = ref(false)
const operating = ref(false)
const saving = ref(false)
const editing = ref(false)
const task = ref({})
const memberList = ref([])
const comments = ref([])
const attachments = ref([])
const activeTab = ref('comments')
const commentText = ref('')
const commenting = ref(false)

// 弹窗状态
const showCloseDialog = ref(false)
const showBlockDialog = ref(false)
const showReviewDialog = ref(false)
const showAiDecompose = ref(false)
const closeEvidence = ref('')
const blockReason = ref('')
const reviewComment = ref('')
const reviewPass = ref(true)
const reviewMode = ref('review') // review | accept

const editForm = reactive({
  title: '', description: '', priority: 2, assignee_id: 0,
  reviewer_id: 0, acceptor_id: 0, due_date: '', estimate: 0,
})

const statusInfo = computed(() => TASK_STATUS_MAP[task.value.status] || { label: '-', type: 'info' })
const priorityInfo = computed(() => PRIORITY_MAP[task.value.priority] || { label: '-', type: 'info' })
const overdue = computed(() => isOverdue(task.value.due_date, task.value.status))
const blockedDays = computed(() => calcBlockedDays(task.value.blocked_since))
const reviewTitle = computed(() => {
  if (reviewMode.value === 'review') return reviewPass.value ? '复核通过' : '复核不通过'
  return reviewPass.value ? '验收通过' : '验收不通过'
})

async function loadData() {
  if (!props.taskId) return
  loading.value = true
  try {
    const [taskRes, commentRes, attachRes] = await Promise.all([
      getTask(props.taskId),
      getComments(props.taskId),
      getAttachments(props.taskId),
    ])
    task.value = taskRes.data || {}
    comments.value = commentRes.data?.list || commentRes.data || []
    attachments.value = attachRes.data?.list || attachRes.data || []
    // 加载项目成员（供编辑选择）
    if (task.value.project_id) {
      const memberRes = await getProjectMembers(task.value.project_id)
      memberList.value = memberRes.data || []
    }
    editing.value = false
  } finally {
    loading.value = false
  }
}

function startEdit() {
  editing.value = true
  editForm.title = task.value.title || ''
  editForm.description = task.value.description || ''
  editForm.priority = task.value.priority || 2
  editForm.assignee_id = task.value.assignee_id || 0
  editForm.reviewer_id = task.value.reviewer_id || 0
  editForm.acceptor_id = task.value.acceptor_id || 0
  editForm.due_date = task.value.due_date ? formatDate(task.value.due_date) : ''
  editForm.estimate = task.value.estimate || 0
}

async function saveEdit() {
  saving.value = true
  try {
    await updateTask(task.value.id, { ...editForm })
    ElMessage.success('保存成功')
    editing.value = false
    await loadData()
    emit('updated')
  } finally {
    saving.value = false
  }
}

async function changeStatus(status) {
  operating.value = true
  try {
    await updateTaskStatus(task.value.id, { status })
    ElMessage.success('状态更新成功')
    await loadData()
    emit('updated')
  } finally {
    operating.value = false
  }
}

async function handleBlock() {
  if (!blockReason.value) {
    ElMessage.warning('请填写阻塞原因')
    return
  }
  operating.value = true
  try {
    await markBlocked(task.value.id, { reason: blockReason.value })
    ElMessage.success('已标记阻塞')
    showBlockDialog.value = false
    blockReason.value = ''
    await loadData()
    emit('updated')
  } finally {
    operating.value = false
  }
}

async function handleUnblock() {
  await ElMessageBox.confirm('确定解除阻塞吗？', '提示', { type: 'warning' })
  operating.value = true
  try {
    await unblockTask(task.value.id)
    ElMessage.success('已解除阻塞')
    await loadData()
    emit('updated')
  } catch {
    // 取消
  } finally {
    operating.value = false
  }
}

function handleReview(pass) {
  reviewMode.value = 'review'
  reviewPass.value = pass
  reviewComment.value = ''
  showReviewDialog.value = true
}

function handleAccept(pass) {
  reviewMode.value = 'accept'
  reviewPass.value = pass
  reviewComment.value = ''
  showReviewDialog.value = true
}

async function submitReview() {
  operating.value = true
  try {
    if (reviewMode.value === 'review') {
      await reviewTask(task.value.id, { pass: reviewPass.value, comment: reviewComment.value })
    } else {
      await acceptTask(task.value.id, { pass: reviewPass.value, comment: reviewComment.value })
    }
    ElMessage.success('操作成功')
    showReviewDialog.value = false
    await loadData()
    emit('updated')
  } finally {
    operating.value = false
  }
}

async function handleCloseTask() {
  if (!closeEvidence.value) {
    ElMessage.warning('请填写关闭证据')
    return
  }
  operating.value = true
  try {
    await closeTask(task.value.id, { evidence: closeEvidence.value })
    ElMessage.success('任务已关闭')
    showCloseDialog.value = false
    closeEvidence.value = ''
    await loadData()
    emit('updated')
  } finally {
    operating.value = false
  }
}

async function submitComment() {
  if (!commentText.value.trim()) return
  commenting.value = true
  try {
    await addComment({ task_id: task.value.id, content: commentText.value })
    commentText.value = ''
    const res = await getComments(task.value.id)
    comments.value = res.data?.list || res.data || []
    ElMessage.success('评论成功')
  } finally {
    commenting.value = false
  }
}

async function handleUpload(options) {
  try {
    await uploadAttachment(options.file, task.value.id)
    ElMessage.success('上传成功')
    const res = await getAttachments(task.value.id)
    attachments.value = res.data?.list || res.data || []
  } catch {
    // 错误已由拦截器处理
  }
}

async function handleDownload(row) {
  const res = await downloadAttachment(row.id)
  downloadBlob(res.data, row.file_name)
}

async function handleDeleteAttachment(row) {
  await ElMessageBox.confirm('确定删除该附件吗？', '提示', { type: 'warning' })
  await deleteAttachment(row.id)
  ElMessage.success('删除成功')
  const res = await getAttachments(task.value.id)
  attachments.value = res.data?.list || res.data || []
}

async function handleAiCommand(cmd) {
  if (cmd === 'decompose') {
    showAiDecompose.value = true
  } else if (cmd === 'generateDesc') {
    const res = await generateTaskDesc({ title: task.value.title, keywords: '' })
    const desc = res.data?.description || res.data?.content || res.data || ''
    if (desc) {
      await ElMessageBox.confirm(desc, 'AI生成的描述', {
        confirmButtonText: '采用',
        cancelButtonText: '取消',
      }).then(async () => {
        await updateTask(task.value.id, { description: typeof desc === 'string' ? desc : JSON.stringify(desc) })
        ElMessage.success('描述已更新')
        await loadData()
        emit('updated')
      }).catch(() => {})
    }
  } else if (cmd === 'suggestAssignee') {
    const res = await suggestAssignee(task.value.id)
    const candidates = res.data?.candidates || res.data || []
    const list = Array.isArray(candidates) ? candidates : [candidates]
    if (list.length) {
      const names = list.map((c) => `${c.nickname || c.name}（待办${c.pending_count || c.pending || 0}个）`).join('、')
      ElMessageBox.confirm(`AI推荐负责人：${names}`, 'AI推荐', {
        confirmButtonText: '分配第一个',
        cancelButtonText: '取消',
      }).then(async () => {
        const first = list[0]
        await updateTask(task.value.id, { assignee_id: first.id })
        ElMessage.success('已分配负责人')
        await loadData()
        emit('updated')
      }).catch(() => {})
    } else {
      ElMessage.info('暂无推荐候选人')
    }
  }
}
</script>

<style lang="scss" scoped>
.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;

  .task-title {
    font-weight: 600;
    font-size: 15px;
  }
}

.task-desc {
  color: #606266;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.blocked-reason {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #f25555;
}

.action-bar {
  margin-top: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.edit-form {
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;

  .edit-actions {
    text-align: right;
  }
}

.comment-list {
  max-height: 240px;
  overflow-y: auto;

  .comment-item {
    display: flex;
    gap: 10px;
    padding: 8px 0;
    border-bottom: 1px solid #f0f0f0;

    .comment-body {
      flex: 1;

      .comment-header {
        display: flex;
        gap: 8px;
        align-items: center;

        .comment-author {
          font-weight: 600;
          font-size: 13px;
        }

        .comment-time {
          font-size: 12px;
          color: #c0c4cc;
        }
      }

      .comment-content {
        font-size: 13px;
        color: #606266;
        margin-top: 2px;
      }
    }
  }
}

.comment-input {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.text-danger {
  color: #f25555;
}
</style>