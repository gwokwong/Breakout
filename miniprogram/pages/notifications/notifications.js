// pages/notifications/notifications.js
const app = getApp()
const { get, post } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    loading: true,
    notifications: [],
    filterTabs: [
      { value: 'all', label: '全部' },
      { value: 'unread', label: '未读' }
    ],
    currentFilter: 'all',
    page: 1,
    hasMore: true
  },

  onShow() {
    if (!app.checkLogin()) return
    this.loadNotifications(true)
  },

  onPullDownRefresh() {
    this.loadNotifications(true).finally(() => wx.stopPullDownRefresh())
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadNotifications(false)
    }
  },

  async loadNotifications(reset) {
    if (reset) {
      this.setData({ page: 1, hasMore: true, loading: true })
    }
    try {
      const params = { page: this.data.page, page_size: 20 }
      const res = await get('/notification', params)
      const list = res.data?.list || res.data || []
      const total = res.data?.total || list.length

      const processed = list.map((n) => {
        const typeInfo = util.NOTIFICATION_TYPE_MAP[n.type] || {}
        return {
          ...n,
          typeLabel: typeInfo || '通知',
          timeText: util.timeAgo(n.created_at)
        }
      })

      // 过滤未读
      let filtered = processed
      if (this.data.currentFilter === 'unread') {
        filtered = processed.filter((n) => !n.is_read)
      }

      const newList = reset ? filtered : [...this.data.notifications, ...filtered]
      this.setData({
        notifications: newList,
        hasMore: newList.length < total,
        page: this.data.page + 1,
        loading: false
      })
    } catch (err) {
      console.error('加载通知失败:', err)
      this.setData({ loading: false })
    }
  },

  onFilterChange(e) {
    this.setData({ currentFilter: e.currentTarget.dataset.value })
    this.loadNotifications(true)
  },

  // 标记单条已读
  async handleRead(e) {
    const id = e.currentTarget.dataset.id
    try {
      await post('/notification/' + id + '/read')
      // 更新本地状态
      const notifications = this.data.notifications.map((n) => {
        if (n.id === id) n.is_read = 1
        return n
      })
      this.setData({ notifications })
    } catch (err) {
      console.error('标记已读失败:', err)
    }
  },

  // 全部已读
  async handleReadAll() {
    try {
      await post('/notification/read-all')
      wx.showToast({ title: '已全部标记为已读', icon: 'success' })
      this.loadNotifications(true)
    } catch (err) {
      console.error('全部已读失败:', err)
    }
  },

  // 点击通知跳转任务
  handleTap(e) {
    const item = e.currentTarget.dataset.item
    // 标记已读
    if (!item.is_read) {
      this.handleRead(e)
    }
    // 跳转任务详情
    if (item.task_id) {
      wx.navigateTo({ url: '/pages/task-detail/task-detail?id=' + item.task_id })
    }
  }
})