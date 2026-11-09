package model

import "time"

// WeeklyReport 周报模型
type WeeklyReport struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TeamID     uint64    `json:"team_id" gorm:"not null;index"`
	ProjectID  uint64    `json:"project_id" gorm:"default:0;index"`
	StartDate  string    `json:"start_date" gorm:"size:20;not null"`
	EndDate    string    `json:"end_date" gorm:"size:20;not null"`
	Content    string    `json:"content" gorm:"type:longtext;not null"`
	GeneratedBy uint64   `json:"generated_by" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// GenerateWeeklyReportReq 生成周报请求
type GenerateWeeklyReportReq struct {
	TeamID    uint64 `json:"team_id"`
	ProjectID uint64 `json:"project_id"`
	StartDate string `json:"start_date" v:"required#开始日期不能为空"`
	EndDate   string `json:"end_date" v:"required#结束日期不能为空"`
}