package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Project 项目控制器
var Project = &projectController{}

type projectController struct{}

// List 获取项目列表
func (c *projectController) List(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	teamID := r.GetCtxVar("team_id").Uint64()

	var projects []model.Project
	query := service.DB.Where("team_id = ?", teamID)

	// 普通成员只能查看自己参与的项目
	userRole := r.GetCtxVar("user_role").Int()
	if userRole == 3 {
		query.Joins("JOIN project_members ON projects.id = project_members.project_id").
			Where("project_members.user_id = ?", userID)
	}

	if err := query.Preload("Owner").Preload("Members").Find(&projects).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取项目列表失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    projects,
	})
}

// Get 获取项目详情
func (c *projectController) Get(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	var project model.Project
	if err := service.DB.Preload("Owner").Preload("Members").Where("id = ?", projectID).First(&project).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取项目详情失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    project,
	})
}

// Create 创建项目
func (c *projectController) Post(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	teamID := r.GetCtxVar("team_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 只有所有者和管理员可以创建项目
	if userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限创建项目",
			"data":    nil,
		})
		return
	}

	var req model.CreateProjectReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 空字符串转nil（datetime列不允许空字符串）
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	project := model.Project{
		Name:        req.Name,
		Description: req.Description,
		TeamID:      teamID,
		OwnerID:     userID,
		Status:      1,
		StartTime:   strPtr(req.StartTime),
		EndTime:     strPtr(req.EndTime),
	}

	// 开启事务
	tx := service.DB.Begin()

	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "创建项目失败",
			"data":    nil,
		})
		return
	}

	// 添加创建者为项目成员
	projectMember := model.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      1, // 项目管理员
	}
	if err := tx.Create(&projectMember).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "添加项目成员失败",
			"data":    nil,
		})
		return
	}

	tx.Commit()

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "创建成功",
		"data":    project,
	})
}

// Update 更新项目
func (c *projectController) Put(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()
	userID := r.GetCtxVar("user_id").Uint64()

	// 检查权限：团队所有者、管理员 或 项目管理员
	var project model.Project
	if err := service.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "项目不存在",
			"data":    nil,
		})
		return
	}

	// 团队级管理员可以修改所有项目
	canEdit := userRole == 1 || userRole == 2
	if !canEdit {
		// 检查是否是项目管理员
		var member model.ProjectMember
		if err := service.DB.Where("project_id = ? AND user_id = ? AND role = 1", projectID, userID).First(&member).Error; err == nil {
			canEdit = true
		}
	}

	if !canEdit {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限修改项目",
			"data":    nil,
		})
		return
	}

	var req model.UpdateProjectReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	data := make(map[string]interface{})
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Status != 0 {
		data["status"] = req.Status
	}
	if req.StartTime != "" {
		data["start_time"] = req.StartTime
	}
	if req.EndTime != "" {
		data["end_time"] = req.EndTime
	}

	if err := service.DB.Model(&model.Project{}).Where("id = ?", projectID).Updates(data).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新项目失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "更新成功",
		"data":    nil,
	})
}

// Delete 删除项目
func (c *projectController) Delete(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()

	// 只有团队所有者和管理员可以删除项目
	if userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限删除项目",
			"data":    nil,
		})
		return
	}

	// 开启事务
	tx := service.DB.Begin()

	// 删除项目
	if err := tx.Delete(&model.Project{}, projectID).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除项目失败",
			"data":    nil,
		})
		return
	}

	// 删除项目成员
	if err := tx.Where("project_id = ?", projectID).Delete(&model.ProjectMember{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除项目成员失败",
			"data":    nil,
		})
		return
	}

	// 项目下的任务project_id置为0
	if err := tx.Model(&model.Task{}).Where("project_id = ?", projectID).Update("project_id", 0).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新任务信息失败",
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

// GetMembers 获取项目成员列表
func (c *projectController) GetMembers(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	var members []model.ProjectMember
	if err := service.DB.Preload("User").
		Where("project_id = ?", projectID).
		Order("role ASC, created_at ASC").
		Find(&members).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取项目成员失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    members,
	})
}

// AddMember 添加项目成员
func (c *projectController) AddMember(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()
	userID := r.GetCtxVar("user_id").Uint64()

	// 检查权限：团队所有者、管理员 或 项目管理员
	canAdd := userRole == 1 || userRole == 2
	if !canAdd {
		var member model.ProjectMember
		if err := service.DB.Where("project_id = ? AND user_id = ? AND role = 1", projectID, userID).First(&member).Error; err == nil {
			canAdd = true
		}
	}

	if !canAdd {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限添加项目成员",
			"data":    nil,
		})
		return
	}

	var req model.AddProjectMemberReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查用户是否存在
	var user model.User
	if err := service.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	// 检查是否已经是项目成员
	var existingMember model.ProjectMember
	if err := service.DB.Where("project_id = ? AND user_id = ?", projectID, req.UserID).First(&existingMember).Error; err == nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户已经是项目成员",
			"data":    nil,
		})
		return
	}

	// 添加成员
	member := model.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	if err := service.DB.Create(&member).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "添加成员失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "添加成功",
		"data":    nil,
	})
}

// RemoveMember 移除项目成员
func (c *projectController) RemoveMember(r *ghttp.Request) {
	projectIDStr := r.Get("id").String()
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的项目ID",
			"data":    nil,
		})
		return
	}

	userRole := r.GetCtxVar("user_role").Int()
	currentUserID := r.GetCtxVar("user_id").Uint64()
	memberIDStr := r.Get("user_id").String()
	memberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的用户ID",
			"data":    nil,
		})
		return
	}

	// 检查权限：团队所有者、管理员 或 项目管理员
	canRemove := userRole == 1 || userRole == 2
	if !canRemove {
		var member model.ProjectMember
		if err := service.DB.Where("project_id = ? AND user_id = ? AND role = 1", projectID, currentUserID).First(&member).Error; err == nil {
			canRemove = true
		}
	}

	if !canRemove {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限移除项目成员",
			"data":    nil,
		})
		return
	}

	// 不能移除自己
	if memberID == currentUserID {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "不能移除自己",
			"data":    nil,
		})
		return
	}

	// 检查成员是否存在
	var member model.ProjectMember
	if err := service.DB.Where("project_id = ? AND user_id = ?", projectID, memberID).First(&member).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户不是项目成员",
			"data":    nil,
		})
		return
	}

	// 不能移除项目管理员（除非是团队所有者）
	if member.Role == 1 && userRole != 1 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限移除项目管理员",
			"data":    nil,
		})
		return
	}

	if err := service.DB.Delete(&member).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "移除成员失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "移除成功",
		"data":    nil,
	})
}