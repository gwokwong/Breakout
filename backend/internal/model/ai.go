package model

import "time"

// AiAction AI操作记录模型
type AiAction struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID     uint64    `json:"task_id" gorm:"not null;index"`
	UserID     uint64    `json:"user_id" gorm:"not null;index"`
	ActionType int       `json:"action_type" gorm:"not null"` // 1-任务拆解 2-周报生成 3-风险评估
	Prompt     string    `json:"prompt" gorm:"type:text;not null"`
	Response   string    `json:"response" gorm:"type:longtext;not null"`
	Status     int       `json:"status" gorm:"default:1"` // 0-失败 1-成功
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// DecomposeTaskReq 任务拆解请求
type DecomposeTaskReq struct {
	TaskID      uint64 `json:"task_id"`
	Title       string `json:"title" v:"required#任务标题不能为空"`
	Description string `json:"description"`
	SubCount    int    `json:"sub_count" d:"3"`
}

// GenerateTaskDescReq 生成任务描述请求
type GenerateTaskDescReq struct {
	Title    string `json:"title" v:"required#任务标题不能为空"`
	Keywords string `json:"keywords"`
}