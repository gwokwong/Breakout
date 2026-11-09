// pages/index/index.js
const app = getApp()
const { get } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    userInfo: null,
    loading: true,
    // 我的任务统计
    myStats: {
      todo: 0,
      inProgress: 0,
      review: 0,
      done: 0
    },
    // 阻塞任务
    blockedTasks: [],
    blockedCount: 0,
    // 项目列表
    projects: [],
    // 最近任务
    recentTasks: [],
    // 未读通知
    unreadCount: 0
  },

  onShow() {
    if (!app.checkLogin()) return
    this.setData({ userInfo: app.globalData.userInfo })
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadData() {
    if (!app.globalData.token) return
    try {
      const userId = app.globalData.userInfo?.id
      const teamId = app.globalData.userInfo?.team_id

      // 并行加载
      const [myTasksRes, blockedRes, projectsRes, unreadRes] = await Promise.all([
        get('/task', { assignee_id: userId, page: 1, page_size: 100 }),
        get('/task/blocked', { page: 1, page_size: 5 }),
        get('/project'),
        get('/notification/unread-count')
      ])

      const myTasks = myTasksRes.data?.list || myTasksRes.data || []
      const blocked = blockedRes.data?.list || blockedRes.data || []
      const projects = projectsRes.data?.list || projectsRes.data || []

      // 统计我的任务
      const stats = { todo: 0, inProgress: 0, review: 0, done: 0 }
      myTasks.forEach((t) => {
        if (t.status === 1) stats.todo++
        else if (t.status === 2) stats.inProgress++
        else if (t.status === 3 || t.status === 4) stats.review++
        else if (t.status === 5 || t.status === 6) stats.done++
      })

      // 处理任务展示数据
      const processTask = (t) => {
        const statusInfo = util.TASK_STATUS_MAP[t.status] || {}
        const priorityInfo = util.PRIORITY_MAP[t.priority] || {}
        return {
          ...t,
          statusLabel: statusInfo.label || '',
          statusColor: statusInfo.color || '#86909c',
          priorityLabel: priorityInfo.label || '',
          priorityColor: priorityInfo.color || '#86909c',
          dueText: util.getDueStatus(t)?.text || '',
          dueColor: util.getDueStatus(t)?.color || '',
          riskScore: util.calcRiskScore(t),
          blockedDays: util.calcBlockedDays(t)
        }
      }

      // 最近任务（按更新时间排序，取前5）
      const recent = [...myTasks]
        .sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at))
        .slice(0, 5)
        .map(processTask)

      this.setData({
        myStats: stats,
        blockedTasks: blocked.map(processTask),
        blockedCount: blockedRes.data?.total || blocked.length,
        projects: projects.slice(0, 3),
        recentTasks: recent,
        unreadCount: unreadRes.data?.count || 0,
        loading: false
      })
    } catch (err) {
      console.error('加载数据失败:', err)
      this.setData({ loading: false })
    }
  },

  // 跳转任务详情
  goTaskDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/task-detail/task-detail?id=' + id })
  },

  // 跳转任务列表
  goTasks() {
    wx.switchTab({ url: '/pages/tasks/tasks' })
  },

  // 跳转通知
  goNotifications() {
    wx.switchTab({ url: '/pages/notifications/notifications' })
  }
})