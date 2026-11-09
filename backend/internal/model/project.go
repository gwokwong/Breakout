package model

import "time"

// Project 项目模型
type Project struct {
	ID          uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TeamID      uint64     `json:"team_id" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	Description string     `json:"description" gorm:"size:1000"`
	Status      int        `json:"status" gorm:"default:1"` // 0-归档 1-进行中 2-暂停
	StartTime   *string    `json:"start_time"`
	EndTime     *string    `json:"end_time"`
	OwnerID     uint64     `json:"owner_id" gorm:"not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联字段
	Owner   *User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Members []ProjectMember `json:"members" gorm:"foreignKey:ProjectID"`
}

// ProjectMember 项目成员模型
type ProjectMember struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID uint64    `json:"project_id" gorm:"not null;uniqueIndex:idx_project_user"`
	UserID    uint64    `json:"user_id" gorm:"not null;uniqueIndex:idx_project_user"`
	Role      int       `json:"role" gorm:"default:2"` // 1-管理员 2-成员
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// 关联字段
	User *User `json:"user" gorm:"foreignKey:UserID"`
}

// CreateProjectReq 创建项目请求
type CreateProjectReq struct {
	Name        string `json:"name" v:"required#项目名称不能为空"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// UpdateProjectReq 更新项目请求
type UpdateProjectReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// AddProjectMemberReq 添加项目成员请求
type AddProjectMemberReq struct {
	UserID uint64 `json:"user_id" v:"required#用户ID不能为空"`
	Role   int    `json:"role" v:"in:1,2#角色值无效"`
}