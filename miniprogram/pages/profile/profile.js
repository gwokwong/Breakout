// pages/profile/profile.js
const app = getApp()
const { get, put } = require('../../utils/request')
const util = require('../../utils/util')

// 比较基础库版本号，v1 >= v2 返回非负数
function compareVersion(v1, v2) {
  const a = String(v1).split('.').map(Number)
  const b = String(v2).split('.').map(Number)
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    const d = (a[i] || 0) - (b[i] || 0)
    if (d !== 0) return d
  }
  return 0
}

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
    // wx.showModal 的 editable 参数需要基础库 2.17.1+，低版本降级提示
    let SDKVersion = '0.0.0'
    try {
      SDKVersion = wx.getSystemInfoSync().SDKVersion || '0.0.0'
    } catch (e) { /* 忽略 */ }
    if (compareVersion(SDKVersion, '2.17.1') < 0) {
      wx.showToast({ title: '当前微信版本过低，请升级微信后修改昵称', icon: 'none' })
      return
    }

    const res = await new Promise((resolve) => {
      wx.showModal({
        title: '修改昵称',
        editable: true,
        placeholderText: '请输入新昵称',
        success: resolve
      })
    })
    if (!res.confirm || !res.content || !res.content.trim()) return

    const nickname = res.content.trim()
    try {
      await put('/user/profile', { nickname })
      // 更新本地
      const userInfo = { ...this.data.userInfo, nickname }
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