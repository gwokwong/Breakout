package model

import "time"

// Team 团队模型
type Team struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	OwnerID     uint64    `json:"owner_id" gorm:"not null"`
	Description string    `json:"description" gorm:"size:500"`
	Status      int       `json:"status" gorm:"default:1"` // 0-禁用 1-正常
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联字段
	Owner   *User  `json:"owner" gorm:"foreignKey:OwnerID"`
	Members []User `json:"members" gorm:"foreignKey:TeamID"`
}

// CreateTeamReq 创建团队请求
type CreateTeamReq struct {
	Name        string `json:"name" v:"required#团队名称不能为空"`
	Description string `json:"description"`
}

// UpdateTeamReq 更新团队请求
type UpdateTeamReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InviteMemberReq 邀请成员请求
type InviteMemberReq struct {
	Phone string `json:"phone" v:"required#手机号不能为空"`
	Role  int    `json:"role" v:"required|in:2,3,4#角色值无效"` // 2-管理员 3-成员 4-访客
}