import request from './request'

// 获取任务列表（支持筛选/排序/分页）
export function getTasks(params) {
  return request.get('/task', { params })
}

// 获取任务树
export function getTaskTree(projectId) {
  return request.get('/task/tree', { params: { project_id: projectId } })
}

// 获取阻塞任务列表
export function getBlockedTasks(params) {
  return request.get('/task/blocked', { params })
}

// 获取任务详情
export function getTask(id) {
  return request.get(`/task/${id}`)
}

// 创建任务
export function createTask(data) {
  return request.post('/task', data)
}

// 更新任务
export function updateTask(id, data) {
  return request.put(`/task/${id}`, data)
}

// 删除任务
export function deleteTask(id) {
  return request.delete(`/task/${id}`)
}

// 更新任务状态
export function updateTaskStatus(id, data) {
  return request.put(`/task/${id}/status`, data)
}

// 添加评论
export function addTaskComment(id, data) {
  return request.post(`/task/${id}/comment`, data)
}

// 标记阻塞
export function markBlocked(id, data) {
  return request.post(`/task/${id}/block`, data)
}

// 解除阻塞
export function unblockTask(id) {
  return request.post(`/task/${id}/unblock`)
}

// 提交复核
export function reviewTask(id, data) {
  return request.post(`/task/${id}/review`, data)
}

// 提交验收
export function acceptTask(id, data) {
  return request.post(`/task/${id}/accept`, data)
}

// 关闭任务
export function closeTask(id, data) {
  return request.post(`/task/${id}/close`, data)
}

// 添加任务依赖关系
export function addTaskRelation(data) {
  return request.post('/task/relation', data)
}

// 删除任务依赖关系
export function deleteTaskRelation(id) {
  return request.delete(`/task/relation/${id}`)
}