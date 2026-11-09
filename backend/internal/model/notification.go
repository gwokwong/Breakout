package model

import "time"

// Notification 通知模型
type Notification struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	TaskID    uint64    `json:"task_id" gorm:"default:0"`
	Type      int       `json:"type" gorm:"not null"` // 1-任务分配 2-任务逾期 3-任务阻塞 4-依赖变更 5-复核请求 6-验收请求 7-周报生成 8-负载过高
	Content   string    `json:"content" gorm:"size:500;not null"`
	IsRead    int       `json:"is_read" gorm:"default:0"` // 0-未读 1-已读
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}