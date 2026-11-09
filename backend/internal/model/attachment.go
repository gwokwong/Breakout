package model

import "time"

// Attachment 附件模型
type Attachment struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uint64    `json:"task_id" gorm:"not null;index"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	FilePath  string    `json:"file_path" gorm:"size:255;not null"`
	FileName  string    `json:"file_name" gorm:"size:255;not null"`
	FileSize  int64     `json:"file_size" gorm:"not null"`
	FileType  string    `json:"file_type" gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联字段
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}