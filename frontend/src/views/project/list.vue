<template>
  <div class="project-list">
    <div class="card-box">
      <div class="flex-between mb-16">
        <h3 class="page-title">项目列表</h3>
        <el-button type="primary" @click="showCreate = true">
          <el-icon><Plus /></el-icon>&nbsp;新建项目
        </el-button>
      </div>

      <el-row :gutter="16" v-loading="loading">
        <el-col v-for="p in projects" :key="p.id" :span="8" class="project-col">
          <el-card shadow="hover" class="project-card" @click="router.push(`/project/${p.id}`)">
            <div class="project-header">
              <el-icon :size="20" color="#4d7bfe"><Folder /></el-icon>
              <span class="project-name">{{ p.name }}</span>
            </div>
            <div class="project-desc">{{ p.description || '暂无描述' }}</div>
            <div class="project-footer">
              <el-tag :type="statusInfo(p.status).type" size="small">{{ statusInfo(p.status).label }}</el-tag>
              <span class="text-muted">{{ formatDate(p.start_time) }} ~ {{ formatDate(p.end_time) }}</span>
            </div>
            <div class="project-actions" @click.stop>
              <el-button link type="primary" size="small" @click="openEdit(p)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(p)">删除</el-button>
            </div>
          </el-card>
        </el-col>
        <el-col :span="24" v-if="!loading && !projects.length">
          <el-empty description="暂无项目，点击右上角创建" />
        </el-col>
      </el-row>
    </div>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="showCreate" :title="editId ? '编辑项目' : '新建项目'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="输入项目名称" />
        </el-form-item>
        <el-form-item label="项目描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="输入项目描述" />
        </el-form-item>
        <el-form-item label="起止时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item v-if="editId" label="项目状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="进行中" :value="1" />
            <el-option label="已暂停" :value="2" />
            <el-option label="已归档" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProjects, createProject, updateProject, deleteProject } from '@/api/project'
import { PROJECT_STATUS_MAP } from '@/constants'
import { formatDate } from '@/utils'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const projects = ref([])
const showCreate = ref(false)
const editId = ref(0)
const formRef = ref()
const dateRange = ref([])

const form = reactive({
  name: '',
  description: '',
  status: 1,
  start_time: '',
  end_time: '',
})

const rules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
}

function statusInfo(status) {
  return PROJECT_STATUS_MAP[status] || { label: '-', type: 'info' }
}

function openEdit(p) {
  editId.value = p.id
  form.name = p.name
  form.description = p.description
  form.status = p.status
  dateRange.value = p.start_time && p.end_time ? [p.start_time, p.end_time] : []
  showCreate.value = true
}

async function handleSave() {
  await formRef.value.validate()
  saving.value = true
  try {
    const data = {
      name: form.name,
      description: form.description,
      start_time: dateRange.value?.[0] || '',
      end_time: dateRange.value?.[1] || '',
    }
    if (editId.value) {
      data.status = form.status
      await updateProject(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createProject(data)
      ElMessage.success('创建成功')
    }
    showCreate.value = false
    editId.value = 0
    await loadProjects()
  } finally {
    saving.value = false
  }
}

async function handleDelete(p) {
  await ElMessageBox.confirm(`确定删除项目「${p.name}」吗？删除后不可恢复`, '警告', { type: 'warning' })
  await deleteProject(p.id)
  ElMessage.success('删除成功')
  await loadProjects()
}

async function loadProjects() {
  loading.value = true
  try {
    const res = await getProjects()
    projects.value = res.data?.list || res.data || []
  } finally {
    loading.value = false
  }
}

// 打开创建弹窗时重置表单
import { watch } from 'vue'
watch(showCreate, (val) => {
  if (!val) {
    editId.value = 0
  } else if (!editId.value) {
    form.name = ''
    form.description = ''
    form.status = 1
    dateRange.value = []
  }
})

onMounted(loadProjects)
</script>

<style lang="scss" scoped>
.page-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.project-col {
  margin-bottom: 16px;
}

.project-card {
  cursor: pointer;

  .project-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;

    .project-name {
      font-size: 16px;
      font-weight: 600;
    }
  }

  .project-desc {
    font-size: 13px;
    color: #909399;
    min-height: 40px;
    margin-bottom: 12px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .project-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .project-actions {
    margin-top: 8px;
    text-align: right;
  }
}
</style>