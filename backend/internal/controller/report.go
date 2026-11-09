package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Report 报表控制器
var Report = &reportController{}

type reportController struct{}

// GetWeeklyList 获取周报列表
func (c *reportController) GetWeeklyList(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	page := r.Get("page", 1).Int()
	pageSize := r.Get("page_size", 20).Int()
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int64
	service.DB.Model(&model.WeeklyReport{}).Where("generated_by = ?", userID).Count(&total)

	var reports []model.WeeklyReport
	if err := service.DB.Where("generated_by = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取周报列表失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"list":      reports,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetWeeklyReport 获取周报详情
func (c *reportController) GetWeeklyReport(r *ghttp.Request) {
	reportIDStr := r.Get("id").String()
	reportID, err := strconv.ParseUint(reportIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的周报ID", "data": nil})
		return
	}

	var report model.WeeklyReport
	if err := service.DB.Where("id = ?", reportID).First(&report).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "周报不存在", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": report})
}

// GenerateWeeklyReport 生成周报
func (c *reportController) GenerateWeeklyReport(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var req model.GenerateWeeklyReportReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 查询该用户在时间范围内的任务
	var tasks []model.Task
	service.DB.Where("assignee_id = ? AND updated_at BETWEEN ? AND ?", userID, req.StartDate+" 00:00:00", req.EndDate+" 23:59:59").
		Find(&tasks)

	// 统计：1-待办 2-进行中 3-待复核 4-待验收 5-已完成 6-已关闭
	var completedCount, inProgressCount, blockedCount int
	var content string
	content = "## 本周工作总结\n\n"

	for _, t := range tasks {
		switch t.Status {
		case 5:
			completedCount++
			content += "- [已完成] " + t.Title + "\n"
		case 2, 3, 4:
			inProgressCount++
			content += "- [进行中] " + t.Title + "\n"
		case 6:
			content += "- [已关闭] " + t.Title + "\n"
		default:
			content += "- [待办] " + t.Title + "\n"
		}
		if t.BlockedReason != "" {
			blockedCount++
			content += "  阻塞原因: " + t.BlockedReason + "\n"
		}
	}

	content += "\n## 统计\n\n"
	content += "- 已完成: " + strconv.Itoa(completedCount) + "\n"
	content += "- 进行中: " + strconv.Itoa(inProgressCount) + "\n"
	content += "- 阻塞: " + strconv.Itoa(blockedCount) + "\n"

	report := model.WeeklyReport{
		TeamID:     req.TeamID,
		ProjectID:  req.ProjectID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Content:    content,
		GeneratedBy: userID,
	}

	if err := service.DB.Create(&report).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "生成周报失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "生成成功",
		"data": g.Map{
			"report":            report,
			"completed_count":   completedCount,
			"in_progress_count": inProgressCount,
			"blocked_count":     blockedCount,
		},
	})
}

// GetStatistics 获取项目统计
func (c *reportController) GetStatistics(r *ghttp.Request) {
	projectID := r.Get("project_id").Uint64()
	if projectID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "项目ID不能为空", "data": nil})
		return
	}

	// 任务状态统计
	type statusCount struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	var statusCounts []statusCount
	service.DB.Model(&model.Task{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&statusCounts)

	// 任务优先级统计
	type priorityCount struct {
		Priority int   `json:"priority"`
		Count    int64 `json:"count"`
	}
	var priorityCounts []priorityCount
	service.DB.Model(&model.Task{}).
		Select("priority, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&priorityCounts)

	// 总任务数
	var totalTasks int64
	service.DB.Model(&model.Task{}).Where("project_id = ?", projectID).Count(&totalTasks)

	// 成员数
	var memberCount int64
	service.DB.Model(&model.ProjectMember{}).Where("project_id = ?", projectID).Count(&memberCount)

	// 已完成任务数（状态5=已完成）
	var completedTasks int64
	service.DB.Model(&model.Task{}).Where("project_id = ? AND status = 5", projectID).Count(&completedTasks)

	// 阻塞任务数（有阻塞原因的任务）
	var blockedTasks int64
	service.DB.Model(&model.Task{}).Where("project_id = ? AND blocked_reason != ''", projectID).Count(&blockedTasks)

	// 完成率
	completionRate := float64(0)
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"total_tasks":     totalTasks,
			"completed_tasks": completedTasks,
			"blocked_tasks":   blockedTasks,
			"member_count":    memberCount,
			"completion_rate": completionRate,
			"by_status":       statusCounts,
			"by_priority":     priorityCounts,
		},
	})
}

// GetGanttChart 获取甘特图数据
func (c *reportController) GetGanttChart(r *ghttp.Request) {
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

	// 构建甘特图数据
	type ganttItem struct {
		ID        uint64  `json:"id"`
		Title     string  `json:"title"`
		ParentID  uint64  `json:"parent_id"`
		Status    int     `json:"status"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		Estimate  float64 `json:"estimate"`
		Actual    float64 `json:"actual"`
		Progress  int     `json:"progress"`
	}

	var items []ganttItem
	for _, t := range tasks {
		item := ganttItem{
			ID:       t.ID,
			Title:    t.Title,
			ParentID: t.ParentID,
			Status:   t.Status,
			Estimate: t.Estimate,
			Actual:   t.Actual,
		}
		if t.StartDate != nil && !t.StartDate.IsZero() {
			item.StartDate = t.StartDate.Format("2006-01-02")
		}
		if t.DueDate != nil && !t.DueDate.IsZero() {
			item.EndDate = t.DueDate.Format("2006-01-02")
		}
		// 计算进度：1-待办 2-进行中 3-待复核 4-待验收 5-已完成 6-已关闭
		switch t.Status {
		case 5, 6:
			item.Progress = 100
		case 4:
			item.Progress = 90
		case 3:
			item.Progress = 75
		case 2:
			item.Progress = 50
		default:
			item.Progress = 0
		}
		items = append(items, item)
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data":    items,
	})
}