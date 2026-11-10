// pages/task-detail/task-detail.js
const app = getApp()
const { get, put, post } = require('../../utils/request')
const util = require('../../utils/util')

Page({
  data: {
    taskId: 0,
    loading: true,
    task: null,
    // 展示数据
    statusLabel: '',
    statusColor: '',
    priorityLabel: '',
    priorityColor: '',
    riskScore: 0,
    riskLevel: {},
    blockedDays: 0,
    dueText: '',
    dueColor: '',
    startDateText: '',
    dueDateText: '',
    // 权限
    canStart: false,      // 开始任务
    canSubmit: false,     // 提交复核
    canReview: false,     // 复核
    canAccept: false,     // 验收
    canClose: false,      // 关闭
    canBlock: false,      // 标记阻塞
    canUnblock: false,    // 解除阻塞
    // 弹窗
    showBlockModal: false,
    blockReason: '',
    showCloseModal: false,
    closeEvidence: '',
    showReviewModal: false,
    reviewComment: '',
    showAcceptModal: false,
    acceptComment: '',
    showCommentModal: false,
    commentContent: '',
    // 评论
    comments: []
  },

  onLoad(options) {
    if (!app.checkLogin()) return
    this.setData({ taskId: options.id })
    this.loadTask()
  },

  onPullDownRefresh() {
    this.loadTask().finally(() => wx.stopPullDownRefresh())
  },

  async loadTask() {
    try {
      const res = await get('/task/' + this.data.taskId)
      const task = res.data
      const userId = app.globalData.userInfo?.id
      const userRole = app.globalData.userInfo?.role

      const statusInfo = util.TASK_STATUS_MAP[task.status] || {}
      const priorityInfo = util.PRIORITY_MAP[task.priority] || {}
      const riskScore = util.calcRiskScore(task)
      const riskLevel = util.getRiskLevel(riskScore)
      const dueStatus = util.getDueStatus(task)

      // 权限判断
      const isAssignee = task.assignee_id === userId
      const isCreator = task.creator_id === userId
      const isReviewer = task.reviewer_id === userId
      const isAcceptor = task.acceptor_id === userId
      const isAdmin = userRole === 1 || userRole === 2
      const isBlocked = !!task.blocked_reason

      this.setData({
        task,
        statusLabel: statusInfo.label || '',
        statusColor: statusInfo.color || '#86909c',
        priorityLabel: priorityInfo.label || '',
        priorityColor: priorityInfo.color || '#86909c',
        riskScore,
        riskLevel,
        blockedDays: util.calcBlockedDays(task),
        dueText: dueStatus?.text || '',
        dueColor: dueStatus?.color || '',
        startDateText: util.formatDate(task.start_date),
        dueDateText: util.formatDate(task.due_date),
        // 权限
        canStart: !isBlocked && task.status === 1 && (isAssignee || isCreator || isAdmin),
        canSubmit: !isBlocked && task.status === 2 && (isAssignee || isCreator || isAdmin),
        canReview: task.status === 3 && isReviewer,
        canAccept: task.status === 4 && isAcceptor,
        canClose: task.status === 5 && isCreator,
        canBlock: !isBlocked && task.status < 5 && (isAssignee || isAdmin),
        canUnblock: isBlocked && (isAssignee || isAdmin),
        loading: false
      })

      // 加载评论
      this.loadComments()
    } catch (err) {
      console.error('加载任务失败:', err)
      this.setData({ loading: false })
    }
  },

  async loadComments() {
    try {
      const res = await get('/comment', { task_id: this.data.taskId, page: 1, page_size: 50 })
      const list = res.data?.list || res.data || []
      const processed = list.map((c) => ({
        ...c,
        timeText: util.timeAgo(c.created_at)
      }))
      this.setData({ comments: processed })
    } catch (err) {
      console.error('加载评论失败:', err)
    }
  },

  // 开始任务（1→2）
  async handleStart() {
    await put('/task/' + this.data.taskId + '/status', { status: 2 })
    wx.showToast({ title: '已开始', icon: 'success' })
    this.loadTask()
  },

  // 提交复核（2→3）
  async handleSubmitReview() {
    await put('/task/' + this.data.taskId + '/status', { status: 3 })
    wx.showToast({ title: '已提交复核', icon: 'success' })
    this.loadTask()
  },

  // 复核通过（3→4）
  async handleReviewPass() {
    const { confirm } = await new Promise((resolve) => {
      wx.showModal({
        title: '确认复核通过',
        content: '复核通过后任务将进入待验收状态，确定吗？',
        success: resolve
      })
    })
    if (!confirm) return
    await post('/task/' + this.data.taskId + '/review', { pass: true, comment: '' })
    wx.showToast({ title: '复核通过', icon: 'success' })
    this.loadTask()
  },

  // 打开复核退回弹窗
  showReviewFail() {
    this.setData({ showReviewModal: true, reviewComment: '' })
  },

  // 复核不通过（3→2）
  async handleReviewFail() {
    const comment = this.data.reviewComment
    if (!comment) {
      wx.showToast({ title: '请填写不通过原因', icon: 'none' })
      return
    }
    await post('/task/' + this.data.taskId + '/review', { pass: false, comment })
    wx.showToast({ title: '已退回', icon: 'success' })
    this.setData({ showReviewModal: false, reviewComment: '' })
    this.loadTask()
  },

  // 验收通过（4→5）
  async handleAcceptPass() {
    const { confirm } = await new Promise((resolve) => {
      wx.showModal({
        title: '确认验收通过',
        content: '验收通过后任务将标记为已完成，确定吗？',
        success: resolve
      })
    })
    if (!confirm) return
    await post('/task/' + this.data.taskId + '/accept', { pass: true, comment: '' })
    wx.showToast({ title: '验收通过', icon: 'success' })
    this.loadTask()
  },

  // 打开验收退回弹窗
  showAcceptFail() {
    this.setData({ showAcceptModal: true, acceptComment: '' })
  },

  // 验收不通过（4→2）
  async handleAcceptFail() {
    const comment = this.data.acceptComment
    if (!comment) {
      wx.showToast({ title: '请填写退回原因', icon: 'none' })
      return
    }
    await post('/task/' + this.data.taskId + '/accept', { pass: false, comment })
    wx.showToast({ title: '已退回', icon: 'success' })
    this.setData({ showAcceptModal: false, acceptComment: '' })
    this.loadTask()
  },

  // 标记阻塞
  showBlock() {
    this.setData({ showBlockModal: true, blockReason: '' })
  },

  async handleBlock() {
    if (!this.data.blockReason) {
      wx.showToast({ title: '请填写阻塞原因', icon: 'none' })
      return
    }
    await post('/task/' + this.data.taskId + '/block', { reason: this.data.blockReason })
    wx.showToast({ title: '已标记阻塞', icon: 'success' })
    this.setData({ showBlockModal: false })
    this.loadTask()
  },

  // 解除阻塞
  async handleUnblock() {
    await post('/task/' + this.data.taskId + '/unblock')
    wx.showToast({ title: '已解除阻塞', icon: 'success' })
    this.loadTask()
  },

  // 关闭任务
  showClose() {
    this.setData({ showCloseModal: true, closeEvidence: '' })
  },

  async handleClose() {
    if (!this.data.closeEvidence) {
      wx.showToast({ title: '请填写关闭证据', icon: 'none' })
      return
    }
    await post('/task/' + this.data.taskId + '/close', { evidence: this.data.closeEvidence })
    wx.showToast({ title: '已关闭', icon: 'success' })
    this.setData({ showCloseModal: false })
    this.loadTask()
  },

  // 评论
  showComment() {
    this.setData({ showCommentModal: true, commentContent: '' })
  },

  async handleComment() {
    if (!this.data.commentContent) {
      wx.showToast({ title: '请输入评论内容', icon: 'none' })
      return
    }
    await post('/task/' + this.data.taskId + '/comment', { content: this.data.commentContent })
    wx.showToast({ title: '评论成功', icon: 'success' })
    this.setData({ showCommentModal: false })
    this.loadComments()
  },

  // 输入
  onBlockReasonInput(e) { this.setData({ blockReason: e.detail.value }) },
  onCloseEvidenceInput(e) { this.setData({ closeEvidence: e.detail.value }) },
  onReviewCommentInput(e) { this.setData({ reviewComment: e.detail.value }) },
  onAcceptCommentInput(e) { this.setData({ acceptComment: e.detail.value }) },
  onCommentInput(e) { this.setData({ commentContent: e.detail.value }) },

  closeModal() {
    this.setData({
      showBlockModal: false,
      showCloseModal: false,
      showReviewModal: false,
      showAcceptModal: false,
      showCommentModal: false
    })
  }
})