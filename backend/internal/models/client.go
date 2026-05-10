package models

import "time"

type Client struct {
	ID                        uint       `gorm:"primarykey;column:id" json:"id"`
	CompanyName               string     `gorm:"column:company_name;size:255;not null" json:"company_name"`
	TodoList                  *string    `gorm:"column:todo_list;type:text" json:"todo_list,omitempty"`
	AccountStatus             string     `gorm:"column:account_status;type:enum('inactive','active','paused');default:'inactive'" json:"account_status"`
	PendingAccountStatus      *string    `gorm:"column:pending_account_status;size:50" json:"pending_account_status,omitempty"`
	PendingStatusRequestedBy  *uint      `gorm:"column:pending_status_requested_by" json:"pending_status_requested_by,omitempty"`
	PendingStatusRequestedAt  *time.Time `gorm:"column:pending_status_requested_at" json:"pending_status_requested_at,omitempty"`
	CreatedBy                 uint       `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt                 time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	Creator     *User        `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	JobRequests []JobRequest `gorm:"foreignKey:ClientID" json:"job_requests,omitempty"`
}

func (Client) TableName() string {
	return "customer_crm"
}
