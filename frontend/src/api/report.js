import request from './request'

// 获取周报列表
export function getWeeklyReports(params) {
  return request.get('/report/weekly', { params })
}

// 获取周报详情
export function getWeeklyReport(id) {
  return request.get(`/report/weekly/${id}`)
}

// 生成周报
export function generateWeeklyReport(data) {
  return request.post('/report/weekly/generate', data)
}

// 获取项目统计
export function getStatistics(projectId) {
  return request.get('/report/statistics', { params: { project_id: projectId } })
}

// 获取甘特图数据
export function getGanttChart(projectId) {
  return request.get('/report/gantt', { params: { project_id: projectId } })
}