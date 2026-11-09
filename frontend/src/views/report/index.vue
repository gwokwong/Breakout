<template>
  <div class="report-page">
    <!-- 工具栏 -->
    <div class="card-box mb-16">
      <div class="flex-between">
        <h3 class="page-title">周报管理</h3>
        <el-button type="primary" @click="showGenerate = true">
          <el-icon><MagicStick /></el-icon>&nbsp;生成周报
        </el-button>
      </div>
    </div>

    <!-- 周报列表 -->
    <div class="card-box">
      <el-table :data="reports" border v-loading="loading" size="small">
        <el-table-column label="报告周期" min-width="180">
          <template #default="{ row }">{{ row.start_date }} ~ {{ row.end_date }}</template>
        </el-table-column>
        <el-table-column label="内容摘要" min-width="300" show-overflow-tooltip>
          <template #default="{ row }">{{ getContentPreview(row.content) }}</template>
        </el-table-column>
        <el-table-column label="生成时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="viewReport(row)">查看</el-button>
            <el-button link type="success" size="small" @click="exportReport(row)">导出</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadReports"
        />
      </div>

      <el-empty v-if="!loading && !reports.length" description="暂无周报，点击右上角生成" />
    </div>

    <!-- 生成周报弹窗 -->
    <el-dialog v-model="showGenerate" title="生成周报" width="520px">
      <el-form ref="genFormRef" :model="genForm" :rules="genRules" label-width="80px">
        <el-form-item label="起止日期" prop="dateRange">
          <el-date-picker
            v-model="genForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="项目">
          <el-select v-model="genForm.project_id" clearable placeholder="全部项目" style="width: 100%">
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-alert type="info" :closable="false" show-icon title="AI将根据选定时间范围内的任务数据自动生成周报" />
      </el-form>
      <template #footer>
        <el-button @click="showGenerate = false">取消</el-button>
        <el-button type="primary" :loading="generating" @click="handleGenerate">生成</el-button>
      </template>
    </el-dialog>

    <!-- 查看周报弹窗 -->
    <el-dialog v-model="showDetail" title="周报详情" width="720px" top="5vh">
      <div v-if="currentReport" class="report-detail">
        <div class="report-header">
          <h3>{{ currentReport.start_date }} ~ {{ currentReport.end_date }}</h3>
          <span class="text-muted">生成时间：{{ formatDateTime(currentReport.created_at) }}</span>
        </div>
        <el-divider />
        <div class="report-content" v-html="renderContent(currentReport.content)"></div>
      </div>
      <template #footer>
        <el-button @click="showDetail = false">关闭</el-button>
        <el-button type="success" @click="exportReport(currentReport)">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getWeeklyReports, generateWeeklyReport } from '@/api/report'
import { getProjects } from '@/api/project'
import { useUserStore } from '@/store/user'
import { formatDateTime, downloadBlob } from '@/utils'

const { state: userState } = useUserStore()

const loading = ref(false)
const generating = ref(false)
const reports = ref([])
const projects = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const showGenerate = ref(false)
const showDetail = ref(false)
const currentReport = ref(null)
const genFormRef = ref()

const genForm = reactive({
  dateRange: [],
  project_id: null,
})

const genRules = {
  dateRange: [{ required: true, message: '请选择起止日期', trigger: 'change' }],
}

function getContentPreview(content) {
  if (!content) return '-'
  // 去除HTML标签取纯文本
  const text = content.replace(/<[^>]+>/g, '')
  return text.length > 80 ? text.slice(0, 80) + '...' : text
}

function renderContent(content) {
  if (!content) return '<p>暂无内容</p>'
  // 如果是JSON格式尝试格式化
  try {
    const parsed = JSON.parse(content)
    return `<pre style="white-space: pre-wrap;">${JSON.stringify(parsed, null, 2)}</pre>`
  } catch {
    // 如果包含HTML标签直接渲染，否则按段落渲染
    if (content.includes('<')) return content
    return content.split('\n').map((line) => `<p>${line}</p>`).join('')
  }
}

function viewReport(row) {
  currentReport.value = row
  showDetail.value = true
}

function exportReport(row) {
  if (!row) return
  const text = `破局周报\n报告周期：${row.start_date} ~ ${row.end_date}\n生成时间：${formatDateTime(row.created_at)}\n\n${row.content?.replace(/<[^>]+>/g, '') || ''}`
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  downloadBlob(blob, `周报_${row.start_date}_${row.end_date}.txt`)
  ElMessage.success('导出成功')
}

async function handleGenerate() {
  await genFormRef.value.validate()
  generating.value = true
  try {
    await generateWeeklyReport({
      start_date: genForm.dateRange[0],
      end_date: genForm.dateRange[1],
      team_id: userState.userInfo?.team_id || 0,
      project_id: genForm.project_id || 0,
    })
    ElMessage.success('周报生成成功')
    showGenerate.value = false
    genForm.dateRange = []
    genForm.project_id = null
    page.value = 1
    await loadReports()
  } finally {
    generating.value = false
  }
}

async function loadReports() {
  loading.value = true
  try {
    const res = await getWeeklyReports({ page: page.value, page_size: pageSize.value })
    reports.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
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
  loadReports()
})
</script>

<style lang="scss" scoped>
.page-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.report-detail {
  .report-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    h3 {
      margin: 0;
      font-size: 18px;
    }
  }

  .report-content {
    max-height: 60vh;
    overflow-y: auto;
    line-height: 1.8;
    font-size: 14px;

    :deep(p) {
      margin: 8px 0;
    }

    :deep(pre) {
      background: #f5f7fa;
      padding: 12px;
      border-radius: 6px;
      font-size: 13px;
    }
  }
}
</style>