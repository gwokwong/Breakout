// app.js
// 根据小程序运行环境自动切换 API 地址
// develop: 开发版（微信开发者工具） trial: 体验版 release: 正式版
const envVersion = (wx.getAccountInfoSync && wx.getAccountInfoSync().miniProgram.envVersion) || 'develop'
const BASE_URLS = {
  develop: 'http://localhost:8080/api/v1',
  trial: 'https://example.com/api/v1',   // TODO: 部署体验版后修改为实际地址
  release: 'https://example.com/api/v1'  // TODO: 部署正式版后修改为实际地址
}

App({
  globalData: {
    userInfo: null,
    token: '',
    baseUrl: BASE_URLS[envVersion] || BASE_URLS.develop
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