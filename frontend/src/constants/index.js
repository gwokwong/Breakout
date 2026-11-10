// ============ 任务状态 ============
export const TASK_STATUS = {
  TODO: 1,      // 待办
  DOING: 2,     // 进行中
  REVIEW: 3,    // 待复核
  ACCEPT: 4,    // 待验收
  DONE: 5,      // 已完成
  CLOSED: 6,    // 已关闭
}

export const TASK_STATUS_MAP = {
  1: { label: '待办', color: '#86909c', type: 'info' },
  2: { label: '进行中', color: '#4d7bfe', type: 'primary' },
  3: { label: '待复核', color: '#f0a020', type: 'warning' },
  4: { label: '待验收', color: '#9B59B6', type: 'warning' },
  5: { label: '已完成', color: '#31c49c', type: 'success' },
  6: { label: '已关闭', color: '#c9cdd4', type: 'info' },
}

export const TASK_STATUS_LIST = Object.entries(TASK_STATUS_MAP).map(([value, item]) => ({
  value: Number(value),
  label: item.label,
}))

// ============ 状态流转规则 ============
// 与后端 validateStatusTransition 保持一致
const STATUS_TRANSITIONS = {
  1: [2, 6],   // 待办 → 进行中/已关闭
  2: [1, 3],   // 进行中 → 待办/待复核
  3: [2],      // 待复核 → 进行中
  4: [2],      // 待验收 → 进行中
  5: [6],      // 已完成 → 已关闭
  6: [],       // 已关闭 → 终态
}

// 校验状态流转是否合法（reviewerId 为0时允许 2→4 跳过复核）
export function canTransition(from, to, reviewerId = 0) {
  if (from === to) return true
  const allowed = STATUS_TRANSITIONS[from] || []
  if (allowed.includes(to)) return true
  // 特例：2→4 仅在无复核人时允许
  if (from === 2 && to === 4 && reviewerId === 0) return true
  return false
}

// ============ 优先级 ============
export const PRIORITY = {
  HIGH: 1,
  MEDIUM: 2,
  LOW: 3,
}

export const PRIORITY_MAP = {
  1: { label: '高', color: '#f25555', type: 'danger' },
  2: { label: '中', color: '#f0a020', type: 'warning' },
  3: { label: '低', color: '#86909c', type: 'info' },
}

export const PRIORITY_LIST = Object.entries(PRIORITY_MAP).map(([value, item]) => ({
  value: Number(value),
  label: item.label,
}))

// ============ 用户角色 ============
export const USER_ROLE = {
  OWNER: 1,   // 所有者
  ADMIN: 2,   // 管理员
  MEMBER: 3,  // 成员
  GUEST: 4,   // 访客
}

export const USER_ROLE_MAP = {
  1: { label: '所有者', type: 'danger' },
  2: { label: '管理员', type: 'warning' },
  3: { label: '成员', type: 'primary' },
  4: { label: '访客', type: 'info' },
}

// ============ 项目状态 ============
export const PROJECT_STATUS_MAP = {
  0: { label: '已归档', type: 'info' },
  1: { label: '进行中', type: 'success' },
  2: { label: '已暂停', type: 'warning' },
}

// ============ 通知类型 ============
export const NOTIFICATION_TYPE_MAP = {
  1: { label: '任务分配', icon: 'User', type: 'primary' },
  2: { label: '任务逾期', icon: 'Warning', type: 'danger' },
  3: { label: '任务阻塞', icon: 'CircleClose', type: 'danger' },
  4: { label: '依赖变更', icon: 'Connection', type: 'warning' },
  5: { label: '复核请求', icon: 'View', type: 'warning' },
  6: { label: '验收请求', icon: 'Select', type: 'warning' },
  7: { label: '周报生成', icon: 'Document', type: 'success' },
  8: { label: '负载过高', icon: 'AlarmClock', type: 'danger' },
  9: { label: '任务评论', icon: 'ChatDotRound', type: 'primary' },
}

// ============ 依赖类型 ============
export const RELATION_TYPE_MAP = {
  FS: '完成-开始（前置完成后，后置才能开始）',
  SS: '开始-开始（前置开始后，后置才能开始）',
  FF: '完成-完成（前置完成后，后置才能完成）',
  SF: '开始-完成（前置开始后，后置才能完成）',
}

export const RELATION_TYPE_LIST = Object.entries(RELATION_TYPE_MAP).map(([value, label]) => ({
  value,
  label,
}))

// ============ 是否里程碑 ============
export const MILESTONE_MAP = {
  0: '否',
  1: '是',
}