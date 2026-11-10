package model

import "time"

// User 用户模型
type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TeamID    uint64    `json:"team_id" gorm:"default:0"`
	Nickname  string    `json:"nickname" gorm:"size:50;not null"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	Phone     *string   `json:"phone,omitempty" gorm:"size:20;uniqueIndex"`
	Email     string    `json:"email" gorm:"size:100"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Role      int       `json:"role" gorm:"default:3"` // 1-所有者 2-管理员 3-成员 4-访客
	Status    int       `json:"status" gorm:"default:1"` // 0-禁用 1-正常
	WxOpenid  *string   `json:"-" gorm:"size:100;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// LoginReq 登录请求
type LoginReq struct {
	Phone    string `json:"phone" v:"required#手机号不能为空"`
	Password string `json:"password" v:"required#密码不能为空"`
}

// RegisterReq 注册请求
type RegisterReq struct {
	Phone    string `json:"phone" v:"required#手机号不能为空"`
	Password string `json:"password" v:"required|min-length:6#密码不能为空|密码长度不能少于6位"`
	Nickname string `json:"nickname" v:"required#昵称不能为空"`
}

// WxLoginReq 微信登录请求
type WxLoginReq struct {
	Code string `json:"code" v:"required#code不能为空"`
}

// UserInfoResp 用户信息响应
type UserInfoResp struct {
	ID       uint64 `json:"id"`
	TeamID   uint64 `json:"team_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
	Role     int    `json:"role"`
	Token    string `json:"token,omitempty"`
}

// UpdateProfileReq 更新用户信息请求（白名单）
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}