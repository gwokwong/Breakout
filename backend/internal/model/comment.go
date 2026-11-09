package model

import "time"

// Comment 评论模型
type Comment struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uint64    `json:"task_id" gorm:"not null;index"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	ParentID  uint64    `json:"parent_id" gorm:"default:0;index"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联字段
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AddCommentReq 添加评论请求
type AddCommentReq struct {
	Content  string `json:"content" v:"required#评论内容不能为空"`
	ParentID uint64 `json:"parent_id"`
}