package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// AI AI辅助控制器
var AI = &aiController{}

type aiController struct{}

// recordAction 记录AI操作
func recordAction(taskID, userID uint64, actionType int, prompt, response string) {
	action := model.AiAction{
		TaskID:     taskID,
		UserID:     userID,
		ActionType: actionType, // 1-任务拆解 2-周报生成 3-风险评估
		Prompt:     prompt,
		Response:   response,
		Status:     1,
	}
	service.DB.Create(&action)
}

// DecomposeTask 任务分解
func (c *aiController) DecomposeTask(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var req model.DecomposeTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 查找父任务
	var parentTask model.Task
	if err := service.DB.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	subCount := req.SubCount
	if subCount <= 0 {
		subCount = 3
	}

	// 基于规则的任务分解：将任务按阶段拆分为子任务
	phases := []struct {
		Suffix      string
		Description string
		Estimate    float64
	}{
		{" - 需求分析", "分析任务需求，明确目标和验收标准", 4},
		{" - 开发实现", "完成核心功能开发", 16},
		{" - 测试验收", "测试功能并提交验收", 4},
		{" - 文档整理", "整理相关文档", 2},
		{" - 上线部署", "部署上线并验证", 2},
	}

	if subCount > len(phases) {
		subCount = len(phases)
	}

	var subTasks []model.Task
	for i := 0; i < subCount; i++ {
		subTask := model.Task{
			TeamID:      parentTask.TeamID,
			ProjectID:   parentTask.ProjectID,
			Title:       parentTask.Title + phases[i].Suffix,
			Description: phases[i].Description,
			Status:      1, // 待办
			Priority:    parentTask.Priority,
			AssigneeID:  parentTask.AssigneeID,
			CreatorID:   userID,
			ParentID:    parentTask.ID,
			SortOrder:   i + 1,
			Estimate:    phases[i].Estimate,
		}
		if err := service.DB.Create(&subTask).Error; err != nil {
			r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "创建子任务失败", "data": nil})
			return
		}
		subTasks = append(subTasks, subTask)
	}

	recordAction(req.TaskID, userID, 1, req.Title, "分解为"+strconv.Itoa(subCount)+"个子任务")

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "任务分解成功",
		"data":    subTasks,
	})
}

// GenerateTaskDescription 生成任务描述
func (c *aiController) GenerateTaskDescription(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var req model.GenerateTaskDescReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 基于规则生成任务描述模板
	description := "## 任务背景\n" + req.Title + "\n\n" +
		"## 目标\n完成" + req.Title + "的开发与交付\n\n"

	if req.Keywords != "" {
		description += "## 关键词\n" + req.Keywords + "\n\n"
	}

	description += "## 验收标准\n- 功能实现完整\n- 代码通过测试\n- 文档已更新\n\n" +
		"## 备注\n请根据项目实际情况补充详细需求。"

	recordAction(0, userID, 3, req.Title, description)

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "生成成功",
		"data": g.Map{
			"description": description,
		},
	})
}

// SuggestAssignee 智能推荐负责人
func (c *aiController) SuggestAssignee(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	var task model.Task
	if err := service.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
		return
	}

	// 查找项目成员，推荐当前任务数最少的成员
	var members []model.User
	service.DB.Table("project_members pm").
		Select("u.*").
		Joins("JOIN users u ON u.id = pm.user_id").
		Where("pm.project_id = ?", task.ProjectID).
		Scan(&members)

	type candidate struct {
		model.User
		TaskCount int64 `json:"task_count"`
	}
	var candidates []candidate
	for _, m := range members {
		var taskCount int64
		// 未完成且未关闭的任务数：状态5=已完成 6=已关闭
		service.DB.Model(&model.Task{}).
			Where("assignee_id = ? AND status NOT IN (5, 6)", m.ID).
			Count(&taskCount)
		candidates = append(candidates, candidate{User: m, TaskCount: taskCount})
	}

	// 按任务数升序排序，任务最少的优先
	if len(candidates) > 1 {
		for i := 0; i < len(candidates)-1; i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].TaskCount < candidates[i].TaskCount {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
	}

	recordAction(taskID, userID, 3, "推荐负责人", "任务ID "+strconv.FormatUint(taskID, 10))

	// 最多返回3个候选人
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    candidates,
	})
}

// GetUsageStats 获取AI使用统计
func (c *aiController) GetUsageStats(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var totalCount int64
	service.DB.Model(&model.AiAction{}).Where("user_id = ?", userID).Count(&totalCount)

	// 按类型统计：1-任务拆解 2-周报生成 3-风险评估
	type typeStat struct {
		ActionType int   `json:"action_type"`
		Count      int64 `json:"count"`
	}
	var stats []typeStat
	service.DB.Model(&model.AiAction{}).
		Select("action_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("action_type").
		Scan(&stats)

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"total_count": totalCount,
			"by_type":     stats,
		},
	})
}