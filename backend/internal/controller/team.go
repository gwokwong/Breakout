package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Team 团队控制器
var Team = &teamController{}

type teamController struct{}

// Get 获取团队信息
func (c *teamController) Get(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()

	var team model.Team
	if err := service.DB.Preload("Owner").Where("id = ?", teamID).First(&team).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "获取团队信息失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    team,
	})
}

// Create 创建团队
func (c *teamController) Post(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var req model.CreateTeamReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	team := model.Team{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		Status:      1,
	}

	if err := service.DB.Create(&team).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "创建团队失败",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "创建成功",
		"data":    team,
	})
}

// Update 更新团队信息
func (c *teamController) Put(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 只有所有者和管理员可以修改团队信息
	if userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限修改团队信息",
			"data":    nil,
		})
		return
	}

	var req model.UpdateTeamReq
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

	if err := service.DB.Model(&model.Team{}).Where("id = ?", teamID).Updates(data).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新失败",
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

// Delete 删除团队
func (c *teamController) Delete(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 只有所有者可以删除团队
	if userRole != 1 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限删除团队",
			"data":    nil,
		})
		return
	}

	// 开启事务
	tx := service.DB.Begin()

	// 删除团队关联数据
	// 1. 删除周报
	if err := tx.Where("team_id = ?", teamID).Delete(&model.WeeklyReport{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除周报失败",
			"data":    nil,
		})
		return
	}
	// 2. 查出团队下所有项目ID
	var projectIDs []uint64
	tx.Model(&model.Project{}).Where("team_id = ?", teamID).Pluck("id", &projectIDs)
	if len(projectIDs) > 0 {
		// 删除项目成员
		if err := tx.Where("project_id IN ?", projectIDs).Delete(&model.ProjectMember{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{
				"code":    http.StatusInternalServerError,
				"message": "删除项目成员失败",
				"data":    nil,
			})
			return
		}
		// 项目下任务的附件、评论、通知、依赖关系
		var taskIDs []uint64
		tx.Model(&model.Task{}).Where("project_id IN ?", projectIDs).Pluck("id", &taskIDs)
		if len(taskIDs) > 0 {
			if err := tx.Where("task_id IN ?", taskIDs).Delete(&model.Attachment{}).Error; err != nil {
				tx.Rollback()
				r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除附件失败", "data": nil})
				return
			}
			if err := tx.Where("task_id IN ?", taskIDs).Delete(&model.Comment{}).Error; err != nil {
				tx.Rollback()
				r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除评论失败", "data": nil})
				return
			}
			if err := tx.Where("task_id IN ?", taskIDs).Delete(&model.Notification{}).Error; err != nil {
				tx.Rollback()
				r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除通知失败", "data": nil})
				return
			}
			if err := tx.Where("predecessor_task_id IN ? OR successor_task_id IN ?", taskIDs, taskIDs).Delete(&model.TaskRelation{}).Error; err != nil {
				tx.Rollback()
				r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除任务依赖失败", "data": nil})
				return
			}
			if err := tx.Where("task_id IN ?", taskIDs).Delete(&model.AiAction{}).Error; err != nil {
				tx.Rollback()
				r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除AI记录失败", "data": nil})
				return
			}
		}
		// 删除任务
		if err := tx.Where("project_id IN ?", projectIDs).Delete(&model.Task{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{
				"code":    http.StatusInternalServerError,
				"message": "删除任务失败",
				"data":    nil,
			})
			return
		}
		// 删除项目
		if err := tx.Where("team_id = ?", teamID).Delete(&model.Project{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{
				"code":    http.StatusInternalServerError,
				"message": "删除项目失败",
				"data":    nil,
			})
			return
		}
	}
	// 4. 删除无项目归属但属于该团队的任务及关联数据
	var orphanTaskIDs []uint64
	tx.Model(&model.Task{}).Where("team_id = ?", teamID).Pluck("id", &orphanTaskIDs)
	if len(orphanTaskIDs) > 0 {
		if err := tx.Where("task_id IN ?", orphanTaskIDs).Delete(&model.Attachment{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除附件失败", "data": nil})
			return
		}
		if err := tx.Where("task_id IN ?", orphanTaskIDs).Delete(&model.Comment{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除评论失败", "data": nil})
			return
		}
		if err := tx.Where("task_id IN ?", orphanTaskIDs).Delete(&model.Notification{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除通知失败", "data": nil})
			return
		}
		if err := tx.Where("predecessor_task_id IN ? OR successor_task_id IN ?", orphanTaskIDs, orphanTaskIDs).Delete(&model.TaskRelation{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除任务依赖失败", "data": nil})
			return
		}
		if err := tx.Where("task_id IN ?", orphanTaskIDs).Delete(&model.AiAction{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除AI记录失败", "data": nil})
			return
		}
		if err := tx.Where("team_id = ?", teamID).Delete(&model.Task{}).Error; err != nil {
			tx.Rollback()
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除任务失败", "data": nil})
			return
		}
	}

	// 删除团队
	if err := tx.Delete(&model.Team{}, teamID).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "删除团队失败",
			"data":    nil,
		})
		return
	}

	// 团队下的用户team_id置为0
	if err := tx.Model(&model.User{}).Where("team_id = ?", teamID).Update("team_id", 0).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "更新用户信息失败",
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

// InviteMember 邀请成员
func (c *teamController) InviteMember(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	// 只有所有者和管理员可以邀请成员
	if userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限邀请成员",
			"data":    nil,
		})
		return
	}

	var req model.InviteMemberReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 查找用户
	var user model.User
	if err := service.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	// 检查用户是否已经在团队中
	if user.TeamID == teamID {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户已经在团队中",
			"data":    nil,
		})
		return
	}

	// 用户已在其他团队，不能直接邀请
	if user.TeamID != 0 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户已在其他团队中，无法邀请",
			"data":    nil,
		})
		return
	}

	// 更新用户团队信息
	user.TeamID = teamID
	user.Role = req.Role
	if err := service.DB.Save(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "邀请失败",
			"data":    nil,
		})
		return
	}

	// TODO: 发送通知
	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "邀请成功",
		"data":    nil,
	})
}

// RemoveMember 移除成员
func (c *teamController) RemoveMember(r *ghttp.Request) {
	teamID := r.GetCtxVar("team_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()
	currentUserID := r.GetCtxVar("user_id").Uint64()

	// 只有所有者和管理员可以移除成员
	if userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限移除成员",
			"data":    nil,
		})
		return
	}

	memberIDStr := r.Get("id").String()
	memberID, err := strconv.ParseUint(memberIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "无效的用户ID",
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

	// 查找用户
	var user model.User
	if err := service.DB.Where("id = ? AND team_id = ?", memberID, teamID).First(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "用户不在团队中",
			"data":    nil,
		})
		return
	}

	// 不能移除所有者
	if user.Role == 1 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusBadRequest,
			"message": "不能移除团队所有者",
			"data":    nil,
		})
		return
	}

	// 管理员不能移除其他管理员
	if userRole == 2 && user.Role == 2 {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusForbidden,
			"message": "没有权限移除管理员",
			"data":    nil,
		})
		return
	}

	// 移除用户出团队
	user.TeamID = 0
	user.Role = 3 // 重置为普通成员角色
	if err := service.DB.Save(&user).Error; err != nil {
		r.Response.WriteJson(g.Map{
			"code":    http.StatusInternalServerError,
			"message": "移除失败",
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