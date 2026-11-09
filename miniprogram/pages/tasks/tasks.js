// pages/tasks/tasks.js
const app = getApp()
const { get } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    loading: true,
    // 筛选
    statusTabs: [
      { value: 0, label: '全部' },
      { value: 1, label: '待办' },
      { value: 2, label: '进行中' },
      { value: 3, label: '待复核' },
      { value: 4, label: '待验收' },
      { value: 5, label: '已完成' },
      { value: 6, label: '已关闭' }
    ],
    currentStatus: 0,
    // 项目筛选
    projects: [],
    currentProject: 0,
    // 任务列表
    tasks: [],
    page: 1,
    hasMore: true
  },

  onShow() {
    if (!app.checkLogin()) return
    this.loadProjects()
    this.loadTasks(true)
  },

  onPullDownRefresh() {
    this.loadTasks(true).finally(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadTasks(false)
    }
  },

  async loadProjects() {
    try {
      const res = await get('/project')
      const list = res.data?.list || res.data || []
      this.setData({
        projects: [{ id: 0, name: '全部项目' }, ...list]
      })
    } catch (err) {
      console.error('加载项目失败:', err)
    }
  },

  async loadTasks(reset) {
    if (reset) {
      this.setData({ page: 1, hasMore: true, loading: true })
    }
    try {
      const params = {
        page: this.data.page,
        page_size: 20
      }
      // 我的任务
      const userId = app.globalData.userInfo?.id
      if (userId) params.assignee_id = userId
      if (this.data.currentStatus > 0) params.status = this.data.currentStatus
      if (this.data.currentProject > 0) params.project_id = this.data.currentProject

      const res = await get('/task', params)
      const list = res.data?.list || res.data || []
      const total = res.data?.total || list.length

      const processed = list.map((t) => {
        const statusInfo = util.TASK_STATUS_MAP[t.status] || {}
        const priorityInfo = util.PRIORITY_MAP[t.priority] || {}
        const dueStatus = util.getDueStatus(t)
        return {
          ...t,
          statusLabel: statusInfo.label || '',
          statusColor: statusInfo.color || '#86909c',
          priorityLabel: priorityInfo.label || '',
          priorityColor: priorityInfo.color || '#86909c',
          dueText: dueStatus?.text || '',
          dueColor: dueStatus?.color || '',
          riskScore: util.calcRiskScore(t),
          blockedDays: util.calcBlockedDays(t),
          estimateText: t.estimate ? t.estimate + 'h' : ''
        }
      })

      const newTasks = reset ? processed : [...this.data.tasks, ...processed]
      this.setData({
        tasks: newTasks,
        hasMore: newTasks.length < total,
        page: this.data.page + 1,
        loading: false
      })
    } catch (err) {
      console.error('加载任务失败:', err)
      this.setData({ loading: false })
    }
  },

  // 切换状态筛选
  onStatusChange(e) {
    const value = e.currentTarget.dataset.value
    this.setData({ currentStatus: value })
    this.loadTasks(true)
  },

  // 切换项目筛选
  onProjectChange(e) {
    const value = e.currentTarget.dataset.value
    this.setData({ currentProject: value })
    this.loadTasks(true)
  },

  // 跳转任务详情
  goTaskDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: '/pages/task-detail/task-detail?id=' + id })
  }
})