<template>
  <el-dialog v-model="visible" title="AI任务拆解" width="640px" @close="handleClose">
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="AI将根据任务信息生成子任务建议，经您确认后才会创建"
      style="margin-bottom: 16px"
    />

    <el-form label-width="90px">
      <el-form-item label="任务标题">
        <el-input v-model="form.title" placeholder="输入任务标题" />
      </el-form-item>
      <el-form-item label="任务描述">
        <el-input v-model="form.description" type="textarea" :rows="3" placeholder="输入任务描述（可选）" />
      </el-form-item>
      <el-form-item label="子任务数">
        <el-slider v-model="form.subCount" :min="2" :max="8" show-input />
      </el-form-item>
    </el-form>

    <div class="ai-actions">
      <el-button type="primary" :loading="generating" @click="handleGenerate">
        <el-icon><MagicStick /></el-icon>&nbsp;生成建议
      </el-button>
    </div>

    <!-- AI生成的建议结果，需用户确认 -->
    <div v-if="suggestions.length" class="suggestion-area">
      <el-divider content-position="left">AI建议的子任务（请确认后创建）</el-divider>
      <div v-for="(item, idx) in suggestions" :key="idx" class="suggestion-item">
        <el-checkbox v-model="item.checked" />
        <span class="suggestion-title">{{ item.title }}</span>
        <span class="suggestion-desc">{{ item.description }}</span>
      </div>
      <div class="confirm-area">
        <el-button type="success" :loading="creating" @click="handleConfirm">
          确认创建选中的子任务（{{ checkedCount }}个）
        </el-button>
      </div>
    </div>

    <!-- 已创建结果 -->
    <div v-if="createdTasks.length" class="created-area">
      <el-divider content-position="left">已创建的子任务</el-divider>
      <el-tag v-for="t in createdTasks" :key="t.id" class="created-tag" type="success">
        {{ t.title }}
      </el-tag>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { decomposeTask, generateTaskDesc } from '@/api/ai'
import { createTask } from '@/api/task'

const props = defineProps({
  taskId: { type: Number, default: 0 },
  projectId: { type: Number, required: true },
  parentId: { type: Number, default: 0 },
  defaultTitle: { type: String, default: '' },
  defaultDesc: { type: String, default: '' },
})

const visible = defineModel('visible', { type: Boolean, default: false })

const generating = ref(false)
const creating = ref(false)
const suggestions = ref([])
const createdTasks = ref([])

const form = reactive({
  title: '',
  description: '',
  subCount: 3,
})

const checkedCount = computed(() => suggestions.value.filter((s) => s.checked).length)

// 打开时初始化
function init() {
  form.title = props.defaultTitle || ''
  form.description = props.defaultDesc || ''
  form.subCount = 3
  suggestions.value = []
  createdTasks.value = []
}

// 生成AI建议（不直接创建）
async function handleGenerate() {
  if (!form.title) {
    ElMessage.warning('请输入任务标题')
    return
  }
  generating.value = true
  try {
    const res = await decomposeTask({
      task_id: props.taskId || props.parentId,
      title: form.title,
      description: form.description,
      sub_count: form.subCount,
    })
    // 解析AI返回的建议列表
    const data = res.data
    let list = []
    if (Array.isArray(data)) {
      list = data
    } else if (data?.sub_tasks) {
      list = data.sub_tasks
    } else if (data?.suggestions) {
      list = data.suggestions
    } else if (typeof data?.response === 'string') {
      // 尝试从文本中解析
      try {
        const parsed = JSON.parse(data.response)
        list = Array.isArray(parsed) ? parsed : parsed.sub_tasks || parsed.suggestions || []
      } catch {
        list = []
      }
    }
    suggestions.value = (list || []).map((item) => ({
      title: item.title || item.name || '',
      description: item.description || item.desc || '',
      checked: true,
    }))
    if (!suggestions.value.length) {
      ElMessage.warning('AI未能生成有效建议，请调整标题后重试')
    } else {
      ElMessage.success(`AI生成了${suggestions.value.length}个子任务建议`)
    }
  } finally {
    generating.value = false
  }
}

// 用户确认后创建选中的子任务
async function handleConfirm() {
  const selected = suggestions.value.filter((s) => s.checked)
  if (!selected.length) {
    ElMessage.warning('请至少选择一个子任务')
    return
  }
  creating.value = true
  try {
    for (const item of selected) {
      const res = await createTask({
        project_id: props.projectId,
        parent_id: props.parentId,
        title: item.title,
        description: item.description,
        priority: 2,
      })
      if (res.data) {
        createdTasks.value.push(res.data)
      }
    }
    ElMessage.success(`成功创建${createdTasks.value.length}个子任务`)
    suggestions.value = []
    emit('created', createdTasks.value)
  } finally {
    creating.value = false
  }
}

const emit = defineEmits(['created'])

function handleClose() {
  suggestions.value = []
  createdTasks.value = []
}

// 监听visible变化初始化
import { watch } from 'vue'
watch(visible, (val) => {
  if (val) init()
})
</script>

<style lang="scss" scoped>
.ai-actions {
  text-align: center;
  margin: 16px 0;
}

.suggestion-area {
  .suggestion-item {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 10px 12px;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    margin-bottom: 8px;

    .suggestion-title {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
      white-space: nowrap;
    }

    .suggestion-desc {
      font-size: 13px;
      color: #86909c;
      flex: 1;
    }
  }

  .confirm-area {
    text-align: center;
    margin-top: 12px;
  }
}

.created-area {
  .created-tag {
    margin: 0 8px 8px 0;
  }
}
</style>