import request from './request'

// 获取项目列表
export function getProjects(params) {
  return request.get('/project', { params })
}

// 创建项目
export function createProject(data) {
  return request.post('/project', data)
}

// 获取项目详情
export function getProject(id) {
  return request.get(`/project/${id}`)
}

// 更新项目
export function updateProject(id, data) {
  return request.put(`/project/${id}`, data)
}

// 删除项目
export function deleteProject(id) {
  return request.delete(`/project/${id}`)
}

// 获取项目成员列表
export function getProjectMembers(id) {
  return request.get(`/project/${id}/members`)
}

// 添加项目成员
export function addProjectMember(id, data) {
  return request.post(`/project/${id}/members`, data)
}

// 移除项目成员
export function removeProjectMember(id, userId) {
  return request.delete(`/project/${id}/member/${userId}`)
}