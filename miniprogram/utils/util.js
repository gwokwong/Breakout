// utils/util.js - 常量与工具函数

// 任务状态映射（与后端/前端一致）
const TASK_STATUS_MAP = {
  1: { label: '待办', color: '#86909c', tagClass: 'tag-info' },
  2: { label: '进行中', color: '#4d7bfe', tagClass: 'tag-primary' },
  3: { label: '待复核', color: '#f0a020', tagClass: 'tag-warning' },
  4: { label: '待验收', color: '#9b59f5', tagClass: 'tag-primary' },
  5: { label: '已完成', color: '#31c49c', tagClass: 'tag-success' },
  6: { label: '已关闭', color: '#c9cdd4', tagClass: 'tag-info' }
}

// 优先级映射
const PRIORITY_MAP = {
  1: { label: '高', color: '#f25555', tagClass: 'tag-danger' },
  2: { label: '中', color: '#f0a020', tagClass: 'tag-warning' },
  3: { label: '低', color: '#31c49c', tagClass: 'tag-success' }
}

// 角色映射
const ROLE_MAP = {
  1: '所有者',
  2: '管理员',
  3: '成员',
  4: '访客'
}

// 通知类型映射
const NOTIFICATION_TYPE_MAP = {
  1: '任务分配',
  2: '任务逾期',
  3: '任务阻塞',
  4: '依赖变更',
  5: '复核请求',
  6: '验收请求',
  7: '周报生成',
  8: '负载过高'
}

// 风险等级
function getRiskLevel(score) {
  if (score >= 60) return { label: '高风险', color: '#f25555', tagClass: 'tag-danger' }
  if (score >= 30) return { label: '中风险', color: '#f0a020', tagClass: 'tag-warning' }
  return { label: '低风险', color: '#31c49c', tagClass: 'tag-success' }
}

// 风险评分计算（与前端 utils/index.js 一致）
function calcRiskScore(task) {
  let score = 0
  const now = new Date()

  // 是否阻塞 +30
  if (task.blocked_reason) score += 30

  // 停滞天数 +5/天 最多+20
  if (task.blocked_since) {
    const days = Math.floor((now - new Date(task.blocked_since)) / 86400000)
    score += Math.min(days * 5, 20)
  }

  // 是否逾期 +25
  if (task.due_date && new Date(task.due_date) < now && task.status < 5) score += 25

  // 距截止剩余天数
  if (task.due_date && task.status < 5) {
    const remainDays = Math.ceil((new Date(task.due_date) - now) / 86400000)
    if (remainDays <= 1) score += 15
    else if (remainDays <= 3) score += 10
    else if (remainDays <= 7) score += 5
  }

  // 优先级
  if (task.priority === 1) score += 5
  else if (task.priority === 2) score += 2

  return score
}

// 格式化日期时间
function formatDateTime(str) {
  if (!str) return ''
  const d = new Date(str)
  if (isNaN(d.getTime())) return str
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 格式化日期
function formatDate(str) {
  if (!str) return ''
  const d = new Date(str)
  if (isNaN(d.getTime())) return str
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 相对时间
function timeAgo(str) {
  if (!str) return ''
  const diff = Date.now() - new Date(str).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return minutes + '分钟前'
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return hours + '小时前'
  const days = Math.floor(hours / 24)
  if (days < 30) return days + '天前'
  return formatDate(str)
}

// 阻塞天数
function calcBlockedDays(task) {
  if (!task.blocked_since) return 0
  return Math.floor((Date.now() - new Date(task.blocked_since).getTime()) / 86400000)
}

// 截止状态
function getDueStatus(task) {
  if (!task.due_date || task.status >= 5) return ''
  const remain = Math.ceil((new Date(task.due_date) - Date.now()) / 86400000)
  if (remain < 0) return { text: '已逾期' + Math.abs(remain) + '天', color: '#f25555' }
  if (remain === 0) return { text: '今天截止', color: '#f0a020' }
  if (remain <= 3) return { text: '剩余' + remain + '天', color: '#f0a020' }
  return { text: '剩余' + remain + '天', color: '#86909c' }
}

module.exports = {
  TASK_STATUS_MAP,
  PRIORITY_MAP,
  ROLE_MAP,
  NOTIFICATION_TYPE_MAP,
  getRiskLevel,
  calcRiskScore,
  formatDateTime,
  formatDate,
  timeAgo,
  calcBlockedDays,
  getDueStatus
}