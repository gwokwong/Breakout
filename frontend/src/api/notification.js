import request from './request'

// 获取通知列表
export function getNotifications(params) {
  return request.get('/notification', { params })
}

// 获取未读通知数量
export function getUnreadCount() {
  return request.get('/notification/unread-count')
}

// 标记所有通知为已读
export function readAllNotifications() {
  return request.post('/notification/read-all')
}

// 清空所有通知
export function clearAllNotifications() {
  return request.post('/notification/clear-all')
}

// 标记单条通知为已读
export function readNotification(id) {
  return request.post(`/notification/${id}/read`)
}

// 删除通知
export function deleteNotification(id) {
  return request.delete(`/notification/${id}`)
}