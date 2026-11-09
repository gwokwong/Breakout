import request from './request'

// AI任务拆解
export function decomposeTask(data) {
  return request.post('/ai/decompose-task', data)
}

// AI生成任务描述
export function generateTaskDesc(data) {
  return request.post('/ai/generate-task-desc', data)
}

// AI推荐负责人
export function suggestAssignee(taskId) {
  return request.post('/ai/suggest-assignee', { task_id: taskId })
}

// AI使用统计
export function getAiUsageStats() {
  return request.get('/ai/usage-stats')
}