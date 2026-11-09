package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Task 任务控制器
var Task = &taskController{}

type taskController struct{}

// List 获取任务列表
func (c *taskController) List(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	teamID := r.GetCtxVar("team_id").Uint64()

	// 可选过滤参数
	projectID := r.Get("project_id").Uint64()
	status := r.Get("status").Int()
	priority := r.Get("priority").Int()
	assigneeID := r.Get("assignee_id").Uint64()
	creatorID := r.Get("creator_id").Uint64()
	dueDateStart := r.Get("due_date_start").String()
	dueDateEnd := r.Get("due_date_end").String()
	keyword := r.Get("keyword").String()

	var tasks []model.Task
	query := service.DB.Where("team_id = ?", teamID)

	// 普通成员只能查看自己参与的项目的任务
	userRole := r.GetCtxVar("user_role").Int()
	if userRole == 3 {
		query.Joins("JOIN project_members ON tasks.project_id = project_members.project_id").
			Where("project_members.user_id = ?", userID)
	}

	// 过滤条件
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if priority > 0 {
		query = query.Where("priority = ?", priority)
	}
	if assigneeID > 0 {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if creatorID > 0 {
		query = query.Where("creator_id = ?", creatorID)
	}
	if dueDateStart != "" {
		query = query.Where("due_date >= ?", dueDateStart)
	}
	if dueDateEnd != "" {
		query = query.Where("due_date <= ?", dueDateEnd)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 排序
	sortBy := r.Get("sort_by", "created_at").String()
	sortOrder := r.Get("sort_order", "desc").String()
	allowedSorts := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"due_date":   true,
		"priority":   true,
		"status":     true,
		"title":      true,
	}
	if allowedSorts[sortBy] {
		if strings.ToLower(sortOrder) == "asc" {
			query = query.Order(sortBy + " ASC")
		} else {
			query = query.Order(sortBy + " DESC")
		}
	}

	// 分页
	page := r.Get("page", 1).Int()
	pageSize := r.Get("page_size", 20).Int()
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.Task{}).Count(&total)

	if err := query.Preload("Project").Preload("Creator").Preload("Assignee").
		Preload("Parent").Preload("Children").Preload("Comments").Preload("Attachments").
		Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取任务列表失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"list":      tasks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Get 获取任务详情
func (c *taskController) Get(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的任务ID",
			"data":    nil,
		})
		return
	}

	var task model.Task
	if err := service.DB.Preload("Project").Preload("Creator").Preload("Assignee").
		Preload("Parent").Preload("Children").Preload("Comments").Preload("Attachments").
		Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取任务详情失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    task,
	})
}

// Create 创建任务
func (c *taskController) Post(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	teamID := r.GetCtxVar("team_id").Uint64()

	var req model.CreateTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查项目是否存在且用户有权限
	var project model.Project
	if err := service.DB.Where("id = ? AND team_id = ?", req.ProjectID, teamID).First(&project).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "项目不存在或无权限",
			"data":    nil,
		})
		return
	}

	task := model.Task{
		Title:       req.Title,
		Description: req.Description,
		TeamID:      teamID,
		ProjectID:   req.ProjectID,
		CreatorID:   userID,
		AssigneeID:  req.AssigneeID,
		ParentID:    req.ParentID,
		Status:      1, // 待开始
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		Estimate:    req.Estimate,
	}

	if err := service.DB.Create(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "创建任务失败",
			"data":    nil,
		})
		return
	}

	// TODO: 发送通知给负责人
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "创建成功",
		"data":    task,
	})
}

// Update 更新任务
func (c *taskController) Put(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的任务ID",
			"data":    nil,
		})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 获取任务信息
	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "任务不存在",
			"data":    nil,
		})
		return
	}

	// 权限检查：团队管理员、任务创建者、任务负责人可以修改
	canEdit := userRole == 1 || userRole == 2 || task.CreatorID == userID || task.AssigneeID == userID
	if !canEdit {
		// 检查是否是项目管理员
		var member model.ProjectMember
		if err := service.DB.Where("project_id = ? AND user_id = ? AND role = 1", task.ProjectID, userID).First(&member).Error; err == nil {
			canEdit = true
		}
	}

	if !canEdit {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限修改任务",
			"data":    nil,
		})
		return
	}

	var req model.UpdateTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	data := make(map[string]interface{})
	if req.Title != "" {
		data["title"] = req.Title
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Status != 0 {
		data["status"] = req.Status
	}
	if req.Priority != 0 {
		data["priority"] = req.Priority
	}
	if req.AssigneeID > 0 {
		data["assignee_id"] = req.AssigneeID
	}
	if req.ProjectID > 0 {
		data["project_id"] = req.ProjectID
	}
	if req.ParentID > 0 {
		data["parent_id"] = req.ParentID
	}
	if req.DueDate != "" {
		data["due_date"] = req.DueDate
	}
	if req.Estimate > 0 {
		data["estimate"] = req.Estimate
	}
	if req.Actual > 0 {
		data["actual"] = req.Actual
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).Updates(data).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新任务失败",
			"data":    nil,
		})
		return
	}

	// TODO: 发送通知给相关人员
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "更新成功",
		"data":    nil,
	})
}

// Delete 删除任务
func (c *taskController) Delete(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的任务ID",
			"data":    nil,
		})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()
	userID := r.GetCtxVar("user_id").Uint64()

	// 获取任务信息
	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "任务不存在",
			"data":    nil,
		})
		return
	}

	// 权限检查：只有团队管理员和任务创建者可以删除
	canDelete := userRole == 1 || userRole == 2 || task.CreatorID == userID
	if !canDelete {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限删除任务",
			"data":    nil,
		})
		return
	}

	// 开启事务
	tx := service.DB.Begin()

	// 删除子任务
	if err := tx.Where("parent_id = ?", taskID).Delete(&model.Task{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除子任务失败",
			"data":    nil,
		})
		return
	}

	// 删除任务评论
	if err := tx.Where("task_id = ?", taskID).Delete(&model.Comment{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除任务评论失败",
			"data":    nil,
		})
		return
	}

	// 删除任务附件
	if err := tx.Where("task_id = ?", taskID).Delete(&model.Attachment{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除任务附件失败",
			"data":    nil,
		})
		return
	}

	// 删除任务本身
	if err := tx.Delete(&model.Task{}, taskID).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除任务失败",
			"data":    nil,
		})
		return
	}

	tx.Commit()

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "删除成功",
		"data":    nil,
	})
}

// UpdateStatus 更新任务状态
func (c *taskController) UpdateStatus(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的任务ID",
			"data":    nil,
		})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 获取任务信息
	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "任务不存在",
			"data":    nil,
		})
		return
	}

	// 权限检查：团队管理员、任务创建者、任务负责人可以修改状态
	canUpdate := userRole == 1 || userRole == 2 || task.CreatorID == userID || task.AssigneeID == userID
	if !canUpdate {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限修改任务状态",
			"data":    nil,
		})
		return
	}

	var req model.UpdateTaskStatusReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	data := map[string]interface{}{
		"status": req.Status,
	}

	if req.Actual > 0 {
		data["actual"] = req.Actual
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).Updates(data).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新任务状态失败",
			"data":    nil,
		})
		return
	}

	// TODO: 发送通知给相关人员
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "更新成功",
		"data":    nil,
	})
}

// AddComment 添加评论
func (c *taskController) AddComment(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的任务ID",
			"data":    nil,
		})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()

	var req model.AddCommentReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	comment := model.Comment{
		TaskID:   taskID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := service.DB.Create(&comment).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "添加评论失败",
			"data":    nil,
		})
		return
	}

	// 预加载用户信息
	service.DB.Preload("User").First(&comment, comment.ID)

	// TODO: 发送通知给相关人员
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "添加成功",
		"data":    comment,
	})
}

// GetTree 获取任务树
func (c *taskController) GetTree(r *ghttp.Request) {
	projectID := r.Get("project_id").Uint64()
	if projectID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "项目ID不能为空", "data": nil})
		return
	}

	var tasks []model.Task
	if err := service.DB.Where("project_id = ?", projectID).
		Order("sort_order ASC, id ASC").
		Find(&tasks).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取任务失败", "data": nil})
		return
	}

	// 构建树结构
	nodeMap := make(map[uint64]*model.TaskTreeNode)
	var roots []model.TaskTreeNode

	// 创建所有节点
	for _, t := range tasks {
		node := &model.TaskTreeNode{
			Task:     t,
			Children: []model.TaskTreeNode{},
		}
		nodeMap[t.ID] = node
	}

	// 构建父子关系
	for _, t := range tasks {
		if t.ParentID == 0 {
			roots = append(roots, *nodeMap[t.ID])
		} else if parent, ok := nodeMap[t.ParentID]; ok {
			parent.Children = append(parent.Children, *nodeMap[t.ID])
		}
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": roots})
}

// GetBlocked 获取阻塞任务列表
func (c *taskController) GetBlocked(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()
	projectID := r.Get("project_id").Uint64()

	query := service.DB.Where("team_id = ? AND blocked_reason != ''", teamID)
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}

	var tasks []model.Task
	if err := query.Preload("Assignee").Preload("Project").
		Order("blocked_since DESC").
		Find(&tasks).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取阻塞任务失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": tasks})
}

// MarkBlocked 标记任务阻塞
func (c *taskController) MarkBlocked(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的任务ID", "data": nil})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	// 权限：负责人或管理员
	userRole := r.GetCtxVar("user_role").Int()
	if task.AssigneeID != userID && userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "只有任务负责人或管理员可以标记阻塞", "data": nil})
		return
	}

	var req model.MarkBlockedReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"blocked_reason": req.Reason,
			"blocked_since":  time.Now(),
		}).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "标记阻塞失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "标记阻塞成功", "data": nil})
}

// MarkUnblocked 解除任务阻塞
func (c *taskController) MarkUnblocked(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的任务ID", "data": nil})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()
	if task.AssigneeID != userID && userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "只有任务负责人或管理员可以解除阻塞", "data": nil})
		return
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"blocked_reason": "",
			"blocked_since":  nil,
		}).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "解除阻塞失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "解除阻塞成功", "data": nil})
}

// Review 复核任务
func (c *taskController) Review(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的任务ID", "data": nil})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	// 只有复核人可以复核
	if task.ReviewerID != userID {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "只有复核人可以复核任务", "data": nil})
		return
	}

	// 任务必须处于待复核状态(3)
	if task.Status != 3 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务必须处于待复核状态", "data": nil})
		return
	}

	var req model.ReviewTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 通过复核 → 待验收(4)；不通过 → 进行中(2)
	newStatus := 4
	if !req.Pass {
		newStatus = 2
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).
		Update("status", newStatus).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "复核失败", "data": nil})
		return
	}

	// 如果有评论，添加为任务评论
	if req.Comment != "" {
		comment := model.Comment{
			TaskID:  taskID,
			UserID:  userID,
			Content: "[复核] " + req.Comment,
		}
		service.DB.Create(&comment)
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "复核成功", "data": g.Map{"status": newStatus}})
}

// Accept 验收任务
func (c *taskController) Accept(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的任务ID", "data": nil})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	// 只有验收人可以验收
	if task.AcceptorID != userID {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "只有验收人可以验收任务", "data": nil})
		return
	}

	// 任务必须处于待验收状态(4)
	if task.Status != 4 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务必须处于待验收状态", "data": nil})
		return
	}

	var req model.AcceptTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 通过验收 → 已完成(5)；不通过 → 进行中(2)
	newStatus := 5
	if !req.Pass {
		newStatus = 2
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).
		Update("status", newStatus).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "验收失败", "data": nil})
		return
	}

	// 如果有评论，添加为任务评论
	if req.Comment != "" {
		comment := model.Comment{
			TaskID:  taskID,
			UserID:  userID,
			Content: "[验收] " + req.Comment,
		}
		service.DB.Create(&comment)
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "验收成功", "data": g.Map{"status": newStatus}})
}

// Close 关闭任务
func (c *taskController) Close(r *ghttp.Request) {
	taskIDStr := r.Get("id").String()
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的任务ID", "data": nil})
		return
	}

	userID := r.GetCtxVar("user_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	// 权限：创建者、管理员可以关闭
	if task.CreatorID != userID && userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "只有任务创建者或管理员可以关闭任务", "data": nil})
		return
	}

	// 任务必须处于已完成状态(5)
	if task.Status != 5 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务必须处于已完成状态才能关闭", "data": nil})
		return
	}

	var req model.CloseTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	if err := service.DB.Model(&model.Task{}).Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":         6, // 已关闭
			"closed_evidence": req.Evidence,
		}).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "关闭任务失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "关闭成功", "data": nil})
}

// AddRelation 添加任务依赖关系
func (c *taskController) AddRelation(r *ghttp.Request) {
	var req model.AddRelationReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 检查前置和后置任务是否存在
	var predecessor model.Task
	if err := service.DB.Where("id = ?", req.PredecessorTaskID).First(&predecessor).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "前置任务不存在", "data": nil})
		return
	}
	var successor model.Task
	if err := service.DB.Where("id = ?", req.SuccessorTaskID).First(&successor).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "后置任务不存在", "data": nil})
		return
	}

	// 不能自依赖
	if req.PredecessorTaskID == req.SuccessorTaskID {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务不能依赖自身", "data": nil})
		return
	}

	// 检查是否已存在
	var count int64
	service.DB.Model(&model.TaskRelation{}).
		Where("predecessor_task_id = ? AND successor_task_id = ?", req.PredecessorTaskID, req.SuccessorTaskID).
		Count(&count)
	if count > 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "依赖关系已存在", "data": nil})
		return
	}

	relation := model.TaskRelation{
		PredecessorTaskID: req.PredecessorTaskID,
		SuccessorTaskID:   req.SuccessorTaskID,
		Type:              req.Type,
	}

	if err := service.DB.Create(&relation).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "添加依赖关系失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "添加成功", "data": relation})
}

// RemoveRelation 删除任务依赖关系
func (c *taskController) RemoveRelation(r *ghttp.Request) {
	relationIDStr := r.Get("id").String()
	relationID, err := strconv.ParseUint(relationIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的依赖关系ID", "data": nil})
		return
	}

	if err := service.DB.Delete(&model.TaskRelation{}, relationID).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除依赖关系失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "删除成功", "data": nil})
}