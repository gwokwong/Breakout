<template>
  <div class="ad-layout">
    <!-- 侧边栏 -->
    <aside class="ad-sidebar" :class="{ 'is-collapsed': isCollapse }">
      <div class="ad-logo">
        <div class="ad-logo-icon">
          <el-icon :size="22"><Promotion /></el-icon>
        </div>
        <transition name="fade">
          <span v-show="!isCollapse" class="ad-logo-text">破局 Breakout</span>
        </transition>
      </div>

      <el-scrollbar class="ad-menu-scroll">
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          class="ad-menu"
        >
          <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>

      <!-- 侧边栏底部折叠按钮 -->
      <div class="ad-sidebar-footer" @click="isCollapse = !isCollapse">
        <el-icon :size="16">
          <Fold v-if="!isCollapse" />
          <Expand v-else />
        </el-icon>
        <span v-show="!isCollapse" class="collapse-text">收起菜单</span>
      </div>
    </aside>

    <!-- 右侧主体 -->
    <div class="ad-main-wrap">
      <!-- 顶栏 -->
      <header class="ad-header">
        <div class="ad-header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentTitle !== '仪表盘'">{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="ad-header-right">
          <!-- 搜索 -->
          <div class="ad-header-search" @click="showSearch = true">
            <el-icon :size="14"><Search /></el-icon>
            <span>搜索</span>
          </div>
          <!-- 全屏 -->
          <div class="ad-header-action" @click="toggleFullscreen">
            <el-icon :size="18">
              <FullScreen v-if="!isFullscreen" />
              <Aim v-else />
            </el-icon>
          </div>
          <!-- 通知 -->
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="ad-header-badge">
            <div class="ad-header-action" @click="router.push('/notifications')">
              <el-icon :size="18"><Bell /></el-icon>
            </div>
          </el-badge>
          <!-- 用户下拉 -->
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="ad-user-info">
              <el-avatar :size="32" :src="userState.userInfo?.avatar" class="ad-avatar">
                {{ userState.userInfo?.nickname?.charAt(0) || '?' }}
              </el-avatar>
              <span class="ad-user-name">{{ userState.userInfo?.nickname || '未登录' }}</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <div class="ad-dropdown-header">
                  <el-avatar :size="40">{{ userState.userInfo?.nickname?.charAt(0) || '?' }}</el-avatar>
                  <div class="ad-dropdown-user">
                    <div class="ad-dropdown-name">{{ userState.userInfo?.nickname }}</div>
                    <div class="ad-dropdown-role">{{ roleLabel }}</div>
                  </div>
                </div>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人资料
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 标签页导航 -->
      <div class="ad-tabs">
        <el-scrollbar>
          <div class="ad-tabs-inner">
            <div
              v-for="tab in visitedTabs"
              :key="tab.path"
              class="ad-tab"
              :class="{ 'is-active': tab.path === route.path }"
              @click="router.push(tab.path)"
            >
              <el-icon :size="12" class="ad-tab-icon"><component :is="tab.icon" /></el-icon>
              <span>{{ tab.title }}</span>
              <el-icon
                v-if="tab.path !== '/dashboard'"
                :size="12"
                class="ad-tab-close"
                @click.stop="closeTab(tab)"
              ><Close /></el-icon>
            </div>
          </div>
        </el-scrollbar>
      </div>

      <!-- 主内容区 -->
      <main class="ad-content">
        <router-view v-slot="{ Component }">
          <transition name="page-transition" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>

    <!-- 搜索弹窗 -->
    <el-dialog v-model="showSearch" width="520px" :show-close="false" class="ad-search-dialog" top="15vh">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索页面..."
        :prefix-icon="Search"
        size="large"
        clearable
        @keyup.enter="handleSearch"
      />
      <div class="ad-search-results">
        <div
          v-for="m in filteredMenus"
          :key="m.path"
          class="ad-search-item"
          @click="router.push(m.path); showSearch = false"
        >
          <el-icon :size="16"><component :is="m.icon" /></el-icon>
          <span>{{ m.title }}</span>
        </div>
        <div v-if="!filteredMenus.length && searchKeyword" class="ad-search-empty">无匹配结果</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { getUnreadCount } from '@/api/notification'

const route = useRoute()
const router = useRouter()
const { state: userState, logout, setTeamInfo } = useUserStore()

const isCollapse = ref(false)
const isFullscreen = ref(false)
const showSearch = ref(false)
const searchKeyword = ref('')
const unreadCount = ref(0)
const visitedTabs = ref([{ path: '/dashboard', title: '仪表盘', icon: 'Odometer' }])
let pollTimer = null

const menuItems = [
  { path: '/dashboard', title: '仪表盘', icon: 'Odometer' },
  { path: '/projects', title: '项目列表', icon: 'Folder' },
  { path: '/tasks', title: '任务列表', icon: 'List' },
  { path: '/kanban', title: '看板', icon: 'Grid' },
  { path: '/gantt', title: '甘特图', icon: 'DataLine' },
  { path: '/blocked', title: '阻塞地图', icon: 'Warning' },
  { path: '/reports', title: '周报', icon: 'Document' },
  { path: '/team', title: '团队管理', icon: 'User' },
  { path: '/notifications', title: '通知中心', icon: 'Bell' },
]

const activeMenu = computed(() => {
  if (route.path.startsWith('/project/')) return '/projects'
  return route.path
})

const currentTitle = computed(() => route.meta?.title || '仪表盘')

const roleLabel = computed(() => {
  const roleMap = { 1: '所有者', 2: '管理员', 3: '成员', 4: '访客' }
  return roleMap[userState.userInfo?.role] || '成员'
})

const filteredMenus = computed(() => {
  if (!searchKeyword.value) return menuItems
  return menuItems.filter((m) => m.title.includes(searchKeyword.value))
})

// 路由变化时添加标签页
watch(() => route.path, (path) => {
  if (path === '/login') return
  const exists = visitedTabs.value.find((t) => t.path === path)
  if (!exists) {
    // 查找菜单图标，默认用 Document
    const menuItem = menuItems.find((m) => {
      if (path.startsWith('/project/')) return m.path === '/projects'
      return m.path === path
    })
    let title = route.meta?.title || '页面'
    // 项目详情页特殊处理
    if (path.startsWith('/project/')) title = '项目面板'
    visitedTabs.value.push({ path, title, icon: menuItem?.icon || 'Document' })
  }
}, { immediate: true })

function closeTab(tab) {
  const idx = visitedTabs.value.findIndex((t) => t.path === tab.path)
  if (idx === -1) return
  visitedTabs.value.splice(idx, 1)
  // 如果关闭的是当前页，跳转到最后一个标签
  if (tab.path === route.path) {
    const last = visitedTabs.value[visitedTabs.value.length - 1]
    if (last) router.push(last.path)
  }
}

function handleSearch() {
  if (filteredMenus.value.length) {
    router.push(filteredMenus.value[0].path)
    showSearch.value = false
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

async function fetchUnreadCount() {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.data?.count || 0
  } catch { /* 静默 */ }
}

function handleCommand(cmd) {
  if (cmd === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', { type: 'warning' }).then(() => {
      logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    }).catch(() => {})
  } else if (cmd === 'profile') {
    ElMessage.info('个人资料功能开发中')
  }
}

onMounted(() => {
  fetchUnreadCount()
  pollTimer = setInterval(fetchUnreadCount, 30000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables.scss' as *;

.ad-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

// ============================================
// 侧边栏
// ============================================
.ad-sidebar {
  width: $ad-sidebar-width;
  background: $ad-bg-sidebar;
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  position: relative;
  z-index: 100;

  &.is-collapsed {
    width: $ad-sidebar-collapsed-width;
  }
}

.ad-logo {
  height: $ad-header-height;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 12px;
  flex-shrink: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);

  .ad-logo-icon {
    width: 36px;
    height: 36px;
    border-radius: $ad-radius;
    background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    flex-shrink: 0;
    box-shadow: 0 2px 8px rgba(77, 123, 254, 0.4);
  }

  .ad-logo-text {
    color: #fff;
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.02em;
    white-space: nowrap;
  }
}

.ad-menu-scroll {
  flex: 1;
  overflow: hidden;
}

.ad-menu {
  border-right: none;
  background: transparent;
  padding: 8px;

  :deep(.el-menu-item) {
    height: 44px;
    line-height: 44px;
    border-radius: $ad-radius;
    margin-bottom: 2px;
    color: rgba(255, 255, 255, 0.65);
    transition: all 0.2s ease;

    &:hover {
      background: rgba(255, 255, 255, 0.08);
      color: #fff;
    }

    &.is-active {
      background: linear-gradient(135deg, #{$ad-primary}, #{$ad-primary-light});
      color: #fff;
      box-shadow: 0 2px 8px rgba(77, 123, 254, 0.35);
    }

    .el-icon {
      font-size: 18px;
    }
  }

  // 折叠状态
  &.el-menu--collapse {
    :deep(.el-menu-item) {
      padding: 0 !important;
      justify-content: center;
    }
  }
}

.ad-sidebar-footer {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.45);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  transition: all 0.2s;
  flex-shrink: 0;

  &:hover {
    color: rgba(255, 255, 255, 0.8);
    background: rgba(255, 255, 255, 0.04);
  }

  .collapse-text {
    font-size: 13px;
    white-space: nowrap;
  }
}

// ============================================
// 右侧主体
// ============================================
.ad-main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

// ============================================
// 顶栏
// ============================================
.ad-header {
  height: $ad-header-height;
  background: $ad-bg-card;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.03);
  z-index: 50;
  flex-shrink: 0;
}

.ad-header-left {
  display: flex;
  align-items: center;

  :deep(.el-breadcrumb__inner) {
    font-weight: 500;
    color: $ad-text-secondary;

    &.is-link:hover {
      color: $ad-primary;
    }
  }

  :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
    color: $ad-text-primary;
    font-weight: 600;
  }
}

.ad-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ad-header-search {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: $ad-bg-page;
  border-radius: 20px;
  cursor: pointer;
  font-size: 13px;
  color: $ad-text-secondary;
  transition: all 0.2s;
  margin-right: 8px;

  &:hover {
    background: $ad-border-light;
    color: $ad-text-regular;
  }
}

.ad-header-action {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: $ad-text-regular;
  transition: all 0.2s;

  &:hover {
    background: $ad-bg-page;
    color: $ad-primary;
  }
}

.ad-header-badge {
  :deep(.el-badge__content) {
    transform: translateY(4px) translateX(6px);
  }
}

.ad-user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px 4px 4px;
  border-radius: 20px;
  transition: background 0.2s;

  &:hover {
    background: $ad-bg-page;
  }

  .ad-avatar {
    background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
    color: #fff;
    font-weight: 600;
  }

  .ad-user-name {
    font-size: 14px;
    font-weight: 500;
    color: $ad-text-primary;
  }
}

// ============================================
// 标签页导航
// ============================================
.ad-tabs {
  height: $ad-tabs-height;
  background: $ad-bg-card;
  border-top: 1px solid $ad-border-light;
  display: flex;
  align-items: center;
  padding: 0 12px;
  flex-shrink: 0;
}

.ad-tabs-inner {
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.ad-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 13px;
  color: $ad-text-secondary;
  cursor: pointer;
  transition: all 0.2s;
  background: transparent;
  border: 1px solid transparent;

  &:hover {
    color: $ad-text-primary;
    background: $ad-bg-page;
  }

  &.is-active {
    color: $ad-primary;
    background: $ad-primary-lighter;
    border-color: rgba(77, 123, 254, 0.15);
    font-weight: 500;
  }

  .ad-tab-close {
    border-radius: 50%;
    padding: 2px;
    transition: all 0.15s;

    &:hover {
      background: rgba(0, 0, 0, 0.1);
      color: $ad-danger;
    }
  }
}

// ============================================
// 主内容区
// ============================================
.ad-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  background: $ad-bg-page;
}

// ============================================
// 搜索弹窗
// ============================================
.ad-search-dialog {
  :deep(.el-dialog) {
    border-radius: $ad-radius-lg;
  }

  :deep(.el-dialog__header) {
    display: none;
  }

  :deep(.el-dialog__body) {
    padding: 16px 20px;
  }

  .ad-search-results {
    margin-top: 12px;
    max-height: 300px;
    overflow-y: auto;

    .ad-search-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 12px;
      border-radius: $ad-radius-sm;
      cursor: pointer;
      color: $ad-text-regular;
      transition: all 0.15s;

      &:hover {
        background: $ad-primary-lighter;
        color: $ad-primary;
      }
    }

    .ad-search-empty {
      text-align: center;
      color: $ad-text-placeholder;
      padding: 24px 0;
      font-size: 13px;
    }
  }
}

// ============================================
// 下拉用户信息
// ============================================
:deep(.ad-dropdown-header) {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px 12px;

  .el-avatar {
    background: linear-gradient(135deg, #{$ad-primary}, #6c8fff);
    color: #fff;
  }

  .ad-dropdown-name {
    font-size: 14px;
    font-weight: 600;
    color: $ad-text-primary;
  }

  .ad-dropdown-role {
    font-size: 12px;
    color: $ad-text-secondary;
  }
}

// 过渡动画
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

.page-transition-enter-active {
  animation: fadeInUp 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.page-transition-leave-active {
  transition: opacity 0.12s ease;
  opacity: 0;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>