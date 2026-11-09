package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Comment 评论控制器
var Comment = &commentController{}

type commentController struct{}

// List 获取评论列表
func (c *commentController) List(r *ghttp.Request) {
	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	page := r.Get("page", 1).Int()
	pageSize := r.Get("page_size", 20).Int()
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int64
	service.DB.Model(&model.Comment{}).Where("task_id = ? AND parent_id = 0", taskID).Count(&total)

	var comments []model.Comment
	if err := service.DB.Preload("User").
		Where("task_id = ? AND parent_id = 0", taskID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&comments).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取评论列表失败", "data": nil})
		return
	}

	// 加载每条评论的回复数
	type commentWithReplies struct {
		model.Comment
		ReplyCount int64 `json:"reply_count"`
	}
	var result []commentWithReplies
	for _, cm := range comments {
		var replyCount int64
		service.DB.Model(&model.Comment{}).Where("parent_id = ?", cm.ID).Count(&replyCount)
		result = append(result, commentWithReplies{Comment: cm, ReplyCount: replyCount})
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"list":      result,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Post 创建评论
func (c *commentController) Post(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var req model.AddCommentReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	comment := model.Comment{
		TaskID:   taskID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := service.DB.Create(&comment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "创建评论失败", "data": nil})
		return
	}

	service.DB.Preload("User").First(&comment, comment.ID)
	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "创建成功", "data": comment})
}

// Delete 删除评论
func (c *commentController) Delete(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	commentIDStr := r.Get("id").String()
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的评论ID", "data": nil})
		return
	}

	var comment model.Comment
	if err := service.DB.Where("id = ?", commentID).First(&comment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "评论不存在", "data": nil})
		return
	}

	// 只有评论作者或管理员可以删除
	if comment.UserID != userID && userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "没有权限删除评论", "data": nil})
		return
	}

	// 删除评论及其回复
	tx := service.DB.Begin()
	if err := tx.Where("parent_id = ?", commentID).Delete(&model.Comment{}).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除回复失败", "data": nil})
		return
	}
	if err := tx.Delete(&model.Comment{}, commentID).Error; err != nil {
		tx.Rollback()
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除评论失败", "data": nil})
		return
	}
	tx.Commit()

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "删除成功", "data": nil})
}

// GetReplies 获取评论回复列表
func (c *commentController) GetReplies(r *ghttp.Request) {
	commentIDStr := r.Get("id").String()
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的评论ID", "data": nil})
		return
	}

	var replies []model.Comment
	if err := service.DB.Preload("User").
		Where("parent_id = ?", commentID).
		Order("created_at ASC").
		Find(&replies).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取回复失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": replies})
}

// CreateReply 创建回复
func (c *commentController) CreateReply(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	commentIDStr := r.Get("id").String()
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的评论ID", "data": nil})
		return
	}

	var req model.AddCommentReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": err.Error(), "data": nil})
		return
	}

	// 检查父评论是否存在
	var parent model.Comment
	if err := service.DB.Where("id = ?", commentID).First(&parent).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "评论不存在", "data": nil})
		return
	}

	reply := model.Comment{
		TaskID:   parent.TaskID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: commentID,
	}

	if err := service.DB.Create(&reply).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "创建回复失败", "data": nil})
		return
	}

	service.DB.Preload("User").First(&reply, reply.ID)
	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "创建成功", "data": reply})
}