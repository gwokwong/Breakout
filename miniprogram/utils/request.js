// utils/request.js - API 请求封装
const app = getApp()

function request(options) {
  const { url, method = 'GET', data = {}, header = {} } = options
  const baseUrl = app.globalData.baseUrl

  return new Promise((resolve, reject) => {
    wx.request({
      url: baseUrl + url,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(app.globalData.token ? { Authorization: 'Bearer ' + app.globalData.token } : {}),
        ...header
      },
      success(res) {
        const body = res.data
        if (res.statusCode === 401 || body.code === 401) {
          // 登录过期
          wx.showToast({ title: '登录已过期', icon: 'none' })
          app.logout()
          reject(new Error('未授权'))
          return
        }
        if (body.code === 200) {
          resolve(body)
        } else {
          wx.showToast({ title: body.message || '请求失败', icon: 'none' })
          reject(new Error(body.message || '请求失败'))
        }
      },
      fail(err) {
        wx.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      }
    })
  })
}

// 便捷方法
function get(url, data) {
  return request({ url, method: 'GET', data })
}

function post(url, data) {
  return request({ url, method: 'POST', data })
}

function put(url, data) {
  return request({ url, method: 'PUT', data })
}

function del(url, data) {
  return request({ url, method: 'DELETE', data })
}

// 文件上传
function upload(url, filePath, formData = {}) {
  const baseUrl = app.globalData.baseUrl
  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: baseUrl + url,
      filePath,
      name: 'file',
      formData,
      header: {
        ...(app.globalData.token ? { Authorization: 'Bearer ' + app.globalData.token } : {})
      },
      success(res) {
        let body
        try {
          body = JSON.parse(res.data)
        } catch (e) {
          wx.showToast({ title: '上传失败：响应格式错误', icon: 'none' })
          reject(new Error('上传响应格式错误'))
          return
        }
        if (body.code === 200) {
          resolve(body)
        } else {
          wx.showToast({ title: body.message || '上传失败', icon: 'none' })
          reject(new Error(body.message || '上传失败'))
        }
      },
      fail(err) {
        wx.showToast({ title: '上传失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

module.exports = { request, get, post, put, del, upload }