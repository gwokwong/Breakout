package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"breakout.cn/backend/internal/model"
	"breakout.cn/backend/internal/service"
)

// Attachment 附件控制器
var Attachment = &attachmentController{}

type attachmentController struct{}

const uploadDir = "./uploads"
const maxUploadSize = 10485760 // 10MB

var allowedFileTypes = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true,
	"pdf": true, "doc": true, "docx": true,
	"xls": true, "xlsx": true, "zip": true, "rar": true,
}

// Upload 上传附件
func (c *attachmentController) Upload(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()

	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	if !checkTaskTeam(r, taskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该任务", "data": nil})
		return
	}

	file := r.GetUploadFile("file")
	if file == nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "请选择要上传的文件", "data": nil})
		return
	}

	// 文件大小校验
	if file.Size > maxUploadSize {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "文件大小不能超过10MB", "data": nil})
		return
	}

	// 文件类型校验
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	if !allowedFileTypes[ext] {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "不支持的文件类型", "data": nil})
		return
	}

	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "创建上传目录失败", "data": nil})
		return
	}

	// 生成唯一文件名（filepath.Base 防止路径穿越攻击）
	safeName := filepath.Base(file.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	dstPath := filepath.Join(uploadDir, filename)

	if _, err := file.Save(filepath.Join(uploadDir, filename)); err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "保存文件失败: " + err.Error(), "data": nil})
		return
	}

	attachment := model.Attachment{
		TaskID:    taskID,
		UserID:    userID,
		FileName:  file.Filename,
		FilePath:  dstPath,
		FileSize:  file.Size,
		FileType:  filepath.Ext(file.Filename),
	}

	if err := service.DB.Create(&attachment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "保存附件记录失败", "data": nil})
		return
	}

	service.DB.Preload("User").First(&attachment, attachment.ID)
	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "上传成功", "data": attachment})
}

// List 获取附件列表
func (c *attachmentController) List(r *ghttp.Request) {
	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	if !checkTaskTeam(r, taskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该任务", "data": nil})
		return
	}

	var attachments []model.Attachment
	if err := service.DB.Preload("User").
		Where("task_id = ?", taskID).
		Order("created_at DESC").
		Find(&attachments).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "获取附件列表失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": attachments})
}

// Download 下载附件
func (c *attachmentController) Download(r *ghttp.Request) {
	attachmentIDStr := r.Get("id").String()
	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的附件ID", "data": nil})
		return
	}

	var attachment model.Attachment
	if err := service.DB.Where("id = ?", attachmentID).First(&attachment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "附件不存在", "data": nil})
		return
	}

	if !checkTaskTeam(r, attachment.TaskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该附件", "data": nil})
		return
	}

	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "文件不存在", "data": nil})
		return
	}

	r.Response.ServeFileDownload(attachment.FilePath, attachment.FileName)
}

// Preview 预览附件
func (c *attachmentController) Preview(r *ghttp.Request) {
	attachmentIDStr := r.Get("id").String()
	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的附件ID", "data": nil})
		return
	}

	var attachment model.Attachment
	if err := service.DB.Where("id = ?", attachmentID).First(&attachment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "附件不存在", "data": nil})
		return
	}

	if !checkTaskTeam(r, attachment.TaskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该附件", "data": nil})
		return
	}

	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "文件不存在", "data": nil})
		return
	}

	r.Response.ServeFile(attachment.FilePath)
}

// Delete 删除附件
func (c *attachmentController) Delete(r *ghttp.Request) {
	userID := r.GetCtxVar("user_id").Uint64()
	userRole := r.GetCtxVar("user_role").Int()

	attachmentIDStr := r.Get("id").String()
	attachmentID, err := strconv.ParseUint(attachmentIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的附件ID", "data": nil})
		return
	}

	var attachment model.Attachment
	if err := service.DB.Where("id = ?", attachmentID).First(&attachment).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusNotFound, "message": "附件不存在", "data": nil})
		return
	}

	if !checkTaskTeam(r, attachment.TaskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该附件", "data": nil})
		return
	}

	// 只有上传者或管理员可以删除
	if attachment.UserID != userID && userRole != 1 && userRole != 2 {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "没有权限删除附件", "data": nil})
		return
	}

	// 删除物理文件
	os.Remove(attachment.FilePath)

	// 删除数据库记录
	if err := service.DB.Delete(&model.Attachment{}, attachmentID).Error; err != nil {
		r.Response.WriteJson(g.Map{"code": http.StatusInternalServerError, "message": "删除附件记录失败", "data": nil})
		return
	}

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "删除成功", "data": nil})
}

// GetStats 获取附件统计
func (c *attachmentController) GetStats(r *ghttp.Request) {
	taskID := r.Get("task_id").Uint64()
	if taskID == 0 {
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "任务ID不能为空", "data": nil})
		return
	}

	if !checkTaskTeam(r, taskID) {
		r.Response.WriteJson(g.Map{"code": http.StatusForbidden, "message": "无权访问该任务", "data": nil})
		return
	}

	var count int64
	var totalSize int64
	service.DB.Model(&model.Attachment{}).Where("task_id = ?", taskID).Count(&count)
	service.DB.Model(&model.Attachment{}).Where("task_id = ?", taskID).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	r.Response.WriteJson(g.Map{"code": http.StatusOK, "message": "success", "data": g.Map{
		"count":      count,
		"total_size": totalSize,
	}})
}