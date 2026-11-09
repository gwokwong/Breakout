import request from './request'

// 上传附件
export function uploadAttachment(file, taskId) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('task_id', taskId)
  return request.post('/attachment/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

// 获取附件列表
export function getAttachments(taskId) {
  return request.get('/attachment', { params: { task_id: taskId } })
}

// 获取附件统计
export function getAttachmentStats(taskId) {
  return request.get('/attachment/stats', { params: { task_id: taskId } })
}

// 下载附件
export function downloadAttachment(id) {
  return request.get(`/attachment/${id}/download`, { responseType: 'blob' })
}

// 预览附件
export function previewAttachment(id) {
  return request.get(`/attachment/${id}/preview`, { responseType: 'blob' })
}

// 删除附件
export function deleteAttachment(id) {
  return request.delete(`/attachment/${id}`)
}