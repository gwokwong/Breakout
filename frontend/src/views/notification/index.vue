<template>
  <div class="notification-page">
    <div class="card-box">
      <div class="flex-between mb-16">
        <h3 class="page-title">通知中心</h3>
        <div class="toolbar">
          <el-button size="small" @click="handleReadAll" :disabled="!unreadCount">全部已读</el-button>
          <el-button size="small" type="danger" plain @click="handleClearAll">清空通知</el-button>
        </div>
      </div>

      <el-tabs v-model="activeTab" @tab-change="loadNotifications">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane :label="`未读 (${unreadCount})`" name="unread" />
      </el-tabs>

      <div v-loading="loading">
        <div v-if="notifications.length" class="notification-list">
          <div
            v-for="n in notifications"
            :key="n.id"
            class="notification-item"
            :class="{ 'is-unread': n.is_read === 0 }"
            @click="handleClickNotification(n)"
          >
            <div class="notification-icon" :style="{ background: getTypeColor(n.type) }">
              <el-icon :size="16" color="#fff"><component :is="getTypeIcon(n.type)" /></el-icon>
            </div>
            <div class="notification-body">
              <div class="notification-content">{{ n.content }}</div>
              <div class="notification-time">{{ formatDateTime(n.created_at) }}</div>
            </div>
            <div class="notification-actions" @click.stop>
              <el-button v-if="n.is_read === 0" link type="primary" size="small" @click="handleRead(n)">已读</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(n)">删除</el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无通知" />

        <div class="pagination-wrap" v-if="total > pageSize">
          <el-pagination
            v-model:current-page="page"
            :page-size="pageSize"
            :total="total"
            layout="prev, pager, next"
            @current-change="loadNotifications"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getNotifications, getUnreadCount, readNotification,
  readAllNotifications, clearAllNotifications, deleteNotification,
} from '@/api/notification'
import { NOTIFICATION_TYPE_MAP } from '@/constants'
import { formatDateTime } from '@/utils'

const router = useRouter()
const loading = ref(false)
const notifications = ref([])
const activeTab = ref('all')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const unreadCount = ref(0)

function getTypeInfo(type) {
  return NOTIFICATION_TYPE_MAP[type] || { label: '通知', icon: 'Bell', type: 'info' }
}

function getTypeIcon(type) {
  return getTypeInfo(type).icon
}

function getTypeColor(type) {
  const colorMap = {
    primary: '#4d7bfe',
    success: '#31c49c',
    warning: '#f0a020',
    danger: '#f25555',
    info: '#909399',
  }
  return colorMap[getTypeInfo(type).type] || '#909399'
}

async function handleClickNotification(n) {
  // 标记已读
  if (n.is_read === 0) {
    await readNotification(n.id)
    n.is_read = 1
    unreadCount.value = Math.max(0, unreadCount.value - 1)
  }
  // 如果关联了任务，跳转到任务详情（通过任务列表页打开）
  if (n.task_id) {
    router.push('/tasks')
  }
}

async function handleRead(n) {
  await readNotification(n.id)
  n.is_read = 1
  unreadCount.value = Math.max(0, unreadCount.value - 1)
  ElMessage.success('已标记为已读')
}

async function handleReadAll() {
  await readAllNotifications()
  ElMessage.success('已全部标记为已读')
  await loadNotifications()
  await fetchUnreadCount()
}

async function handleClearAll() {
  await ElMessageBox.confirm('确定清空所有通知吗？', '警告', { type: 'warning' })
  await clearAllNotifications()
  ElMessage.success('已清空')
  await loadNotifications()
  await fetchUnreadCount()
}

async function handleDelete(n) {
  await deleteNotification(n.id)
  ElMessage.success('已删除')
  await loadNotifications()
  await fetchUnreadCount()
}

async function loadNotifications() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (activeTab.value === 'unread') params.is_read = 0
    const res = await getNotifications(params)
    notifications.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

async function fetchUnreadCount() {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data?.count || 0
  } catch { /* 忽略 */ }
}

onMounted(() => {
  loadNotifications()
  fetchUnreadCount()
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
  gap: 8px;
}

.notification-list {
  .notification-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 8px;
    border-bottom: 1px solid #f0f0f0;
    cursor: pointer;
    border-radius: 4px;

    &:hover {
      background: #f5f7fa;
    }

    &.is-unread {
      background: #ecf5ff;

      &:hover {
        background: #d9ecff;
      }

      .notification-content {
        font-weight: 600;
      }
    }

    .notification-icon {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    .notification-body {
      flex: 1;
      min-width: 0;

      .notification-content {
        font-size: 14px;
        color: #303133;
        line-height: 1.5;
      }

      .notification-time {
        font-size: 12px;
        color: #c0c4cc;
        margin-top: 4px;
      }
    }

    .notification-actions {
      flex-shrink: 0;
      display: flex;
      gap: 4px;
    }
  }
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
</style>