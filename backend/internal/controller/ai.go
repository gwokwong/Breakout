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

// DecomposeTask 任务分解（仅生成建议，不直接创建；用户确认后由前端调用 CreateTask 创建）
func (c *aiController) DecomposeTask(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	teamID := r.GetCtxVar("team_id").Uint64()

	var req model.DecomposeTaskReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 基准标题：优先使用请求中的标题，其次取父任务标题
	baseTitle := req.Title
	priority := 2
	if req.TaskID > 0 {
		var parentTask model.Task
		if err := service.DB.Where("id = ?", req.TaskID).First(&parentTask).Error; err != nil {
			r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "任务不存在", "data": nil})
			return
		}
		// 校验任务归属团队
		if parentTask.TeamID != teamID {
			r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权操作此任务", "data": nil})
			return
		}
		if baseTitle == "" {
			baseTitle = parentTask.Title
		}
		priority = parentTask.Priority
	}
	if baseTitle == "" {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务标题不能为空", "data": nil})
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

	// 仅生成建议列表，不写库（用户确认后由前端调用 CreateTask 创建）
	suggestions := make([]g.Map, 0, subCount)
	for i := 0; i < subCount; i++ {
		suggestions = append(suggestions, g.Map{
			"title":       baseTitle + phases[i].Suffix,
			"description": phases[i].Description,
			"estimate":    phases[i].Estimate,
			"priority":    priority,
		})
	}

	recordAction(req.TaskID, userID, 1, baseTitle, "生成"+strconv.Itoa(subCount)+"个子任务建议")

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "生成成功",
		"data": g.Map{
			"sub_tasks": suggestions,
		},
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
	teamID := r.GetCtxVar("team_id").Uint64()

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

	// 校验任务归属团队
	if task.TeamID != teamID {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权操作此任务", "data": nil})
		return
	}

	// 查找项目成员及未完成任务数（单次JOIN查询避免N+1）
	type candidate struct {
		model.User
		TaskCount int64 `json:"task_count"`
	}
	var candidates []candidate
	service.DB.Table("project_member pm").
		Select("u.*, COUNT(t.id) as task_count").
		Joins("JOIN user u ON u.id = pm.user_id").
		Joins("LEFT JOIN task t ON t.assignee_id = u.id AND t.status NOT IN (5, 6)").
		Where("pm.project_id = ?", task.ProjectID).
		Group("u.id").
		Order("task_count ASC").
		Scan(&candidates)

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