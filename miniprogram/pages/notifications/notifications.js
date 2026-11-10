// pages/notifications/notifications.js
const app = getApp()
const { get, post } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    loading: true,
    notifications: [],
    allNotifications: [],
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
    if (this.data.loading && !reset) return
    if (reset) {
      this.setData({ page: 1, hasMore: true })
    }
    this.setData({ loading: true })
    const seq = (this._loadSeq = (this._loadSeq || 0) + 1)
    try {
      const params = { page: this.data.page, page_size: 20 }
      const res = await get('/notification', params)
      if (seq !== this._loadSeq) return
      const list = res.data?.list || res.data || []
      const total = res.data?.total || list.length

      const processed = list.map((n) => ({
        ...n,
        typeLabel: util.NOTIFICATION_TYPE_MAP[n.type] || '通知',
        timeText: util.timeAgo(n.created_at)
      }))

      const allList = reset ? processed : [...this.data.allNotifications, ...processed]
      const displayed = this.data.currentFilter === 'unread'
        ? allList.filter((n) => !n.is_read)
        : allList

      this.setData({
        allNotifications: allList,
        notifications: displayed,
        hasMore: allList.length < total,
        page: this.data.page + 1,
        loading: false
      })
    } catch (err) {
      console.error('加载通知失败:', err)
      if (seq === this._loadSeq) this.setData({ loading: false })
    }
  },

  onFilterChange(e) {
    const currentFilter = e.currentTarget.dataset.value
    const displayed = currentFilter === 'unread'
      ? this.data.allNotifications.filter((n) => !n.is_read)
      : this.data.allNotifications
    this.setData({ currentFilter, notifications: displayed })
  },

  // 标记单条已读
  async handleRead(e) {
    const id = e.currentTarget.dataset.id
    try {
      await post('/notification/' + id + '/read')
      // 更新本地状态
      const allNotifications = this.data.allNotifications.map((n) =>
        n.id === id ? { ...n, is_read: 1 } : n
      )
      const displayed = this.data.currentFilter === 'unread'
        ? allNotifications.filter((n) => !n.is_read)
        : allNotifications
      this.setData({ allNotifications, notifications: displayed })
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