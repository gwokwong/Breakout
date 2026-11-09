package model

import "time"

// Task 任务模型
type Task struct {
	ID             uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TeamID         uint64     `json:"team_id" gorm:"not null;index"`
	ProjectID      uint64     `json:"project_id" gorm:"not null;index"`
	ParentID       uint64     `json:"parent_id" gorm:"default:0;index"`
	Title          string     `json:"title" gorm:"size:255;not null"`
	Description    string     `json:"description" gorm:"type:text"`
	Status         int        `json:"status" gorm:"default:1"`   // 1-待办 2-进行中 3-待复核 4-待验收 5-已完成 6-已关闭
	Priority       int        `json:"priority" gorm:"default:2"` // 1-高 2-中 3-低
	AssigneeID     uint64     `json:"assignee_id" gorm:"default:0"`
	ReviewerID     uint64     `json:"reviewer_id" gorm:"default:0"`
	AcceptorID     uint64     `json:"acceptor_id" gorm:"default:0"`
	CreatorID      uint64     `json:"creator_id" gorm:"not null"`
	DueDate        *time.Time `json:"due_date"`
	StartDate      *time.Time `json:"start_date"`
	Estimate       float64    `json:"estimate" gorm:"default:0"` // 预估工时（小时）
	Actual         float64    `json:"actual" gorm:"default:0"`   // 实际工时（小时）
	BlockedReason  string     `json:"blocked_reason" gorm:"size:500"`
	BlockedSince   *time.Time `json:"blocked_since"`
	ClosedEvidence string     `json:"closed_evidence" gorm:"type:text"`
	SortOrder      int        `json:"sort_order" gorm:"default:0"`
	IsMilestone    int        `json:"is_milestone" gorm:"default:0"` // 0-否 1-是
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联字段
	Project     *Project     `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Assignee    *User        `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
	Reviewer    *User        `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
	Acceptor    *User        `json:"acceptor,omitempty" gorm:"foreignKey:AcceptorID"`
	Creator     *User        `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Parent      *Task        `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Task       `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Comments    []Comment    `json:"comments,omitempty" gorm:"foreignKey:TaskID"`
	Attachments []Attachment `json:"attachments,omitempty" gorm:"foreignKey:TaskID"`
}

// TaskRelation 任务依赖关系
type TaskRelation struct {
	ID                uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	PredecessorTaskID uint64    `json:"predecessor_task_id" gorm:"not null;uniqueIndex:idx_pre_suc"`
	SuccessorTaskID   uint64    `json:"successor_task_id" gorm:"not null;uniqueIndex:idx_pre_suc"`
	Type              string    `json:"type" gorm:"size:2;default:'FS'"` // FS/SS/FF/SF
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CreateTaskReq 创建任务请求
type CreateTaskReq struct {
	ProjectID   uint64     `json:"project_id" v:"required#项目ID不能为空"`
	ParentID    uint64     `json:"parent_id"`
	Title       string     `json:"title" v:"required#任务标题不能为空"`
	Description string     `json:"description"`
	Priority    int        `json:"priority" v:"in:1,2,3#优先级值无效"`
	AssigneeID  uint64     `json:"assignee_id"`
	ReviewerID  uint64     `json:"reviewer_id"`
	AcceptorID  uint64     `json:"acceptor_id"`
	DueDate     *time.Time `json:"due_date"`
	StartDate   *time.Time `json:"start_date"`
	Estimate    float64    `json:"estimate"`
	IsMilestone int        `json:"is_milestone" v:"in:0,1#是否为里程碑值无效"`
}

// UpdateTaskReq 更新任务请求
type UpdateTaskReq struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      int     `json:"status" v:"in:1,2,3,4,5,6#状态值无效"`
	Priority    int     `json:"priority" v:"in:1,2,3#优先级值无效"`
	AssigneeID  uint64  `json:"assignee_id"`
	ReviewerID  uint64  `json:"reviewer_id"`
	AcceptorID  uint64  `json:"acceptor_id"`
	ProjectID   uint64  `json:"project_id"`
	ParentID    uint64  `json:"parent_id"`
	DueDate     string  `json:"due_date"`
	StartDate   string  `json:"start_date"`
	Estimate    float64 `json:"estimate"`
	Actual      float64 `json:"actual"`
	SortOrder   int     `json:"sort_order"`
}

// UpdateTaskStatusReq 更新任务状态请求
type UpdateTaskStatusReq struct {
	Status int     `json:"status" v:"required|in:1,2,3,4,5,6#状态不能为空|状态值无效"`
	Actual float64 `json:"actual"`
}

// MarkBlockedReq 标记阻塞请求
type MarkBlockedReq struct {
	Reason string `json:"reason" v:"required#阻塞原因不能为空"`
}

// ReviewTaskReq 复核任务请求
type ReviewTaskReq struct {
	Pass    bool   `json:"pass"`
	Comment string `json:"comment"`
}

// AcceptTaskReq 验收任务请求
type AcceptTaskReq struct {
	Pass    bool   `json:"pass"`
	Comment string `json:"comment"`
}

// CloseTaskReq 关闭任务请求
type CloseTaskReq struct {
	Evidence string `json:"evidence" v:"required#关闭证据不能为空"`
}

// AddRelationReq 添加任务依赖请求
type AddRelationReq struct {
	PredecessorTaskID uint64 `json:"predecessor_task_id" v:"required#前置任务ID不能为空"`
	SuccessorTaskID   uint64 `json:"successor_task_id" v:"required#后置任务ID不能为空"`
	Type              string `json:"type" v:"in:FS,SS,FF,SF#依赖类型无效"`
}

// TaskTreeNode 任务树节点
type TaskTreeNode struct {
	Task
	Progress     float64         `json:"progress"`
	TotalSub     int             `json:"total_sub"`
	CompletedSub int             `json:"completed_sub"`
	Children     []TaskTreeNode  `json:"children"`
}