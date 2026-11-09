package controller

import (
	"net/http"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Notification 通知控制器
var Notification = &notificationController{}

type notificationController struct{}

// List 获取通知列表
func (c *notificationController) List(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	page := r.Get("page", 1).Int()
	pageSize := r.Get("page_size", 20).Int()
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var total int64
	service.DB.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total)

	var notifications []model.Notification
	if err := service.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&notifications).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取通知列表失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    http.StatusOK,
		"message": "success",
		"data": g.Map{
			"list":      notifications,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUnreadCount 获取未读通知数量
func (c *notificationController) GetUnreadCount(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	var count int64
	service.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = 0", userID).Count(&count)

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": g.Map{"count": count}})
}

// MarkRead 标记单条通知为已读
func (c *notificationController) MarkRead(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	notificationIDStr := r.Get("id").String()
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的通知ID", "data": nil})
		return
	}

	if err := service.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", 1).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "标记已读失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "标记成功", "data": nil})
}

// MarkAllRead 标记所有通知为已读
func (c *notificationController) MarkAllRead(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	if err := service.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Update("is_read", 1).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "标记全部已读失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "标记成功", "data": nil})
}

// Delete 删除通知
func (c *notificationController) Delete(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	notificationIDStr := r.Get("id").String()
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的通知ID", "data": nil})
		return
	}

	if err := service.DB.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&model.Notification{}).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除通知失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "删除成功", "data": nil})
}

// ClearAll 清空所有通知
func (c *notificationController) ClearAll(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	if err := service.DB.Where("user_id = ?", userID).
		Delete(&model.Notification{}).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "清空通知失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "清空成功", "data": nil})
}