// pages/profile/profile.js
const app = getApp()
const { get, put } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    userInfo: null,
    roleLabel: '',
    teamInfo: null,
    myTaskCount: 0,
    unreadCount: 0
  },

  onShow() {
    if (!app.checkLogin()) return
    this.setData({
      userInfo: app.globalData.userInfo,
      roleLabel: util.ROLE_MAP[app.globalData.userInfo?.role] || '成员'
    })
    this.loadStats()
  },

  async loadStats() {
    try {
      const userId = app.globalData.userInfo?.id
      const [taskRes, unreadRes, teamRes] = await Promise.all([
        get('/task', { assignee_id: userId, status: 2, page: 1, page_size: 1 }),
        get('/notification/unread-count'),
        get('/team')
      ])

      this.setData({
        myTaskCount: taskRes.data?.total || 0,
        unreadCount: unreadRes.data?.count || 0,
        teamInfo: teamRes.data
      })
    } catch (err) {
      console.error('加载统计失败:', err)
    }
  },

  // 编辑昵称
  async handleEditNickname() {
    const res = await new Promise((resolve) => {
      wx.showModal({
        title: '修改昵称',
        editable: true,
        placeholderText: '请输入新昵称',
        success: resolve
      })
    })
    if (!res.confirm || !res.content) return

    try {
      await put('/user/profile', { nickname: res.content })
      // 更新本地
      const userInfo = { ...this.data.userInfo, nickname: res.content }
      app.globalData.userInfo = userInfo
      wx.setStorageSync('userInfo', userInfo)
      this.setData({ userInfo })
      wx.showToast({ title: '修改成功', icon: 'success' })
    } catch (err) {
      console.error('修改昵称失败:', err)
    }
  },

  // 退出登录
  handleLogout() {
    wx.showModal({
      title: '提示',
      content: '确定退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout()
        }
      }
    })
  }
})