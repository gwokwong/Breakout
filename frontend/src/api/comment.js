import request from './request'

// 获取评论列表
export function getComments(taskId) {
  return request.get('/comment', { params: { task_id: taskId } })
}

// 添加评论
export function addComment(data) {
  return request.post('/comment', data)
}

// 删除评论
export function deleteComment(id) {
  return request.delete(`/comment/${id}`)
}

// 获取评论回复
export function getReplies(id) {
  return request.get(`/comment/${id}/replies`)
}

// 回复评论
export function replyComment(id, data) {
  return request.post(`/comment/${id}/reply`, data)
}