import request from './request'

// 获取团队信息
export function getTeam() {
  return request.get('/team')
}

// 创建团队
export function createTeam(data) {
  return request.post('/team', data)
}

// 更新团队
export function updateTeam(data) {
  return request.put('/team', data)
}

// 删除团队
export function deleteTeam() {
  return request.delete('/team')
}

// 邀请成员
export function inviteMember(data) {
  return request.post('/team/invite', data)
}

// 移除成员
export function removeMember(userId) {
  return request.delete(`/team/member/${userId}`)
}