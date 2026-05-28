package models

import "time"

const (
	TaskVerifyDetails       = "verify_details"
	TaskBusinessFlow        = "business_flow"
	TaskCRM                 = "crm"
	TaskBusinessAccelerator = "business_accelerator"
	TaskDatabaseReactive    = "database_reactive"
	TaskOnboarding          = "onboarding"
)

var AllTaskTypes = []string{
	TaskVerifyDetails,
	TaskBusinessFlow,
	TaskCRM,
	TaskBusinessAccelerator,
	TaskDatabaseReactive,
	TaskOnboarding,
}

type Task struct {
	ID           uint       `gorm:"primarykey;column:id" json:"id"`
	JobRequestID uint       `gorm:"column:job_request_id;not null" json:"job_request_id"`
	TaskType     string     `gorm:"column:task_type;type:enum('verify_details','business_flow','crm','business_accelerator','database_reactive','onboarding');not null" json:"task_type"`
	Status       string     `gorm:"column:status;type:enum('pending','in_progress','pending_on_client','completed');default:'pending'" json:"status"`
	Remarks      *string    `gorm:"column:remarks;type:text" json:"remarks,omitempty"`
	UpdatedBy    *uint      `gorm:"column:updated_by" json:"updated_by,omitempty"`
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	UpdatedByUser *User `gorm:"foreignKey:UpdatedBy" json:"updated_by_user,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}
