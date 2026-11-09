// pages/login/login.js
const app = getApp()
const { post } = require('../../utils/request')

Page({
  data: {
    loading: false,
    showPhoneLogin: false
  },

  onLoad() {
    // 已登录直接跳首页
    if (app.globalData.token) {
      wx.switchTab({ url: '/pages/index/index' })
    }
  },

  // 微信一键登录
  async handleWxLogin() {
    if (this.data.loading) return
    this.setData({ loading: true })

    try {
      // 1. 调用 wx.login 获取 code
      const loginRes = await new Promise((resolve, reject) => {
        wx.login({
          success: resolve,
          fail: reject
        })
      })

      if (!loginRes.code) {
        wx.showToast({ title: '获取微信登录凭证失败', icon: 'none' })
        return
      }

      // 2. 发送 code 到后端换取 token
      const res = await post('/user/wx-login', { code: loginRes.code })
      app.saveLogin(res.data)
      wx.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' })
      }, 500)
    } catch (err) {
      console.error('微信登录失败:', err)
      // 微信登录失败时，回退到手机号登录表单
      this.setData({ showPhoneLogin: true })
    } finally {
      this.setData({ loading: false })
    }
  },

  // 手机号登录
  async handlePhoneLogin(e) {
    const { phone, password } = e.detail.value
    if (!phone || !password) {
      wx.showToast({ title: '请输入手机号和密码', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    try {
      const res = await post('/user/login', { phone, password })
      app.saveLogin(res.data)
      wx.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        wx.switchTab({ url: '/pages/index/index' })
      }, 500)
    } catch (err) {
      console.error('登录失败:', err)
    } finally {
      this.setData({ loading: false })
    }
  }
})