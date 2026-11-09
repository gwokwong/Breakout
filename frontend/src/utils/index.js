import dayjs from 'dayjs'
import { TASK_STATUS } from '@/constants'

// ============ 日期格式化 ============
export function formatDate(date, fmt = 'YYYY-MM-DD') {
  if (!date) return '-'
  return dayjs(date).format(fmt)
}

export function formatDateTime(date) {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

// ============ 文件大小格式化 ============
export function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(1)} ${units[i]}`
}

// ============ 计算停滞天数（阻塞持续天数） ============
export function calcBlockedDays(blockedSince) {
  if (!blockedSince) return 0
  return dayjs().diff(dayjs(blockedSince), 'day')
}

// ============ 计算距截止日期剩余天数 ============
export function calcDaysLeft(dueDate) {
  if (!dueDate) return null
  return dayjs(dueDate).startOf('day').diff(dayjs().startOf('day'), 'day')
}

// ============ 是否逾期 ============
export function isOverdue(dueDate, status) {
  if (!dueDate) return false
  // 已完成/已关闭不算逾期
  if (status === TASK_STATUS.DONE || status === TASK_STATUS.CLOSED) return false
  return dayjs(dueDate).isBefore(dayjs(), 'day')
}

// ============ 风险分计算 ============
// 规则（PRD）：是否阻塞、停滞天数、是否逾期、后置任务影响数、负责人负载、距截止剩余天数
export function calcRiskScore(task) {
  let score = 0
  // 是否阻塞 (+30)
  if (task.blocked_reason) {
    score += 30
    // 停滞天数：每天+5，最多+20
    const days = calcBlockedDays(task.blocked_since)
    score += Math.min(days * 5, 20)
  }
  // 是否逾期 (+25)
  if (isOverdue(task.due_date, task.status)) {
    score += 25
  }
  // 后置任务影响数（有多少任务依赖此任务）(+8/个，最多+16)
  if (task.successor_count) {
    score += Math.min(task.successor_count * 8, 16)
  }
  // 负责人负载（负责人未完成任务数）(+2/个，最多+10)
  if (task.assignee_load) {
    score += Math.min(task.assignee_load * 2, 10)
  }
  // 距截止剩余天数：<=1天+15，<=3天+10，<=7天+5
  const daysLeft = calcDaysLeft(task.due_date)
  if (daysLeft !== null && task.status !== TASK_STATUS.DONE && task.status !== TASK_STATUS.CLOSED) {
    if (daysLeft <= 1) score += 15
    else if (daysLeft <= 3) score += 10
    else if (daysLeft <= 7) score += 5
  }
  // 优先级加权：高+5 中+2 低+0
  if (task.priority === 1) score += 5
  else if (task.priority === 2) score += 2
  return score
}

// ============ 风险等级 ============
export function getRiskLevel(score) {
  if (score >= 60) return { label: '高风险', color: '#f25555', type: 'danger' }
  if (score >= 30) return { label: '中风险', color: '#f0a020', type: 'warning' }
  return { label: '低风险', color: '#31c49c', type: 'success' }
}

// ============ 任务状态进度百分比（与后端GetGanttChart一致） ============
export function statusToProgress(status) {
  const map = {
    [TASK_STATUS.DONE]: 100,
    [TASK_STATUS.CLOSED]: 100,
    [TASK_STATUS.ACCEPT]: 90,
    [TASK_STATUS.REVIEW]: 75,
    [TASK_STATUS.DOING]: 50,
    [TASK_STATUS.TODO]: 0,
  }
  return map[status] || 0
}

// ============ 下载Blob文件 ============
export function downloadBlob(blob, filename) {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}