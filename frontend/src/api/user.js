import request from './request'

// 用户登录
export function login(data) {
  return request.post('/user/login', data)
}

// 用户注册
export function register(data) {
  return request.post('/user/register', data)
}

// 微信登录
export function wxLogin(data) {
  return request.post('/user/wx-login', data)
}

// 获取当前用户信息
export function getUserInfo() {
  return request.get('/user/info')
}

// 更新个人资料
export function updateProfile(data) {
  return request.put('/user/profile', data)
}