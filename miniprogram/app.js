// app.js
App({
  globalData: {
    userInfo: null,
    token: '',
    // 开发环境地址，部署后改为线上服务器地址
    baseUrl: 'http://localhost:8080/api/v1'
  },

  onLaunch() {
    // 从本地缓存恢复登录态
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')
    if (token) {
      this.globalData.token = token
      this.globalData.userInfo = userInfo
    }
  },

  // 检查登录态，未登录跳转登录页
  checkLogin() {
    if (!this.globalData.token) {
      wx.redirectTo({ url: '/pages/login/login' })
      return false
    }
    return true
  },

  // 保存登录信息
  saveLogin(data) {
    this.globalData.token = data.token
    this.globalData.userInfo = {
      id: data.id,
      team_id: data.team_id,
      nickname: data.nickname,
      avatar: data.avatar,
      phone: data.phone,
      role: data.role
    }
    wx.setStorageSync('token', data.token)
    wx.setStorageSync('userInfo', this.globalData.userInfo)
  },

  // 退出登录
  logout() {
    this.globalData.token = ''
    this.globalData.userInfo = null
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
    wx.redirectTo({ url: '/pages/login/login' })
  }
})