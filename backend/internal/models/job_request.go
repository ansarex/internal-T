package models

import "time"

type JobRequest struct {
	ID               uint       `gorm:"primarykey;column:id" json:"id"`
	ClientID         uint       `gorm:"column:client_id;not null" json:"client_id"`
	Status           string     `gorm:"column:status;type:enum('pending','client_pending','pending_to_owner','completed');default:'pending'" json:"status"`
	CurrentStage     int        `gorm:"column:current_stage;default:1" json:"current_stage"`
	Indicator        string     `gorm:"column:indicator;type:enum('grey','yellow','green','red');default:'grey'" json:"indicator"`
	CustomerPIC          *string    `gorm:"column:customer_pic;size:255" json:"customer_pic,omitempty"`
	MonthlyRecurring     *float64   `gorm:"column:monthly_recurring;type:decimal(10,2)" json:"monthly_recurring,omitempty"`
	AccountType          *string    `gorm:"column:account_type;size:255" json:"account_type,omitempty"`
	RecurringStartDate   *time.Time `gorm:"column:recurring_start_date;type:date" json:"recurring_start_date,omitempty"`
	AssignedSalesID  *uint      `gorm:"column:assigned_sales_id" json:"assigned_sales_id,omitempty"`
	AssignedCSID     *uint      `gorm:"column:assigned_cs_id" json:"assigned_cs_id,omitempty"`
	SignedFilePath    *string    `gorm:"column:signed_file_path;size:255" json:"signed_file_path,omitempty"`
	SignedUploadedAt  *time.Time `gorm:"column:signed_uploaded_at" json:"signed_uploaded_at,omitempty"`
	SignedUploadedBy  *uint      `gorm:"column:signed_uploaded_by" json:"signed_uploaded_by,omitempty"`
	SLAStartedAt     *time.Time `gorm:"column:sla_started_at" json:"sla_started_at,omitempty"`
	SLADeadline      *time.Time `gorm:"column:sla_deadline" json:"sla_deadline,omitempty"`
	LastActivityAt   *time.Time `gorm:"column:last_activity_at" json:"last_activity_at,omitempty"`
	Stage1ApprovedAt *time.Time `gorm:"column:stage1_approved_at" json:"stage1_approved_at,omitempty"`
	Stage1ApprovedBy *uint      `gorm:"column:stage1_approved_by" json:"stage1_approved_by,omitempty"`
	CreatedBy        uint       `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	Client         *Client     `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	AssignedSales  *User       `gorm:"foreignKey:AssignedSalesID" json:"assigned_sales,omitempty"`
	AssignedCS     *User       `gorm:"foreignKey:AssignedCSID" json:"assigned_cs,omitempty"`
	SignedUploader *User       `gorm:"foreignKey:SignedUploadedBy" json:"signed_uploader,omitempty"`
	Stage1Approver *User       `gorm:"foreignKey:Stage1ApprovedBy" json:"stage1_approver,omitempty"`
	Creator        *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Tasks          []Task      `gorm:"foreignKey:JobRequestID" json:"tasks,omitempty"`
	Agreements     []Agreement `gorm:"foreignKey:JobRequestID" json:"agreements,omitempty"`
}

func (JobRequest) TableName() string {
	return "job_requests"
}
