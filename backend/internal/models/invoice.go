package models

import "time"

type Invoice struct {
	ID              uint       `gorm:"primarykey;column:id" json:"id"`
	InvoiceNumber   string     `gorm:"column:invoice_number;size:255;not null;uniqueIndex" json:"invoice_number"`
	ClientID        uint       `gorm:"column:client_id;not null" json:"client_id"`
	JobRequestID    uint       `gorm:"column:job_request_id;not null" json:"job_request_id"`
	AssignedSalesID *uint      `gorm:"column:assigned_sales_id" json:"assigned_sales_id,omitempty"`
	AssignedCSID    *uint      `gorm:"column:assigned_cs_id" json:"assigned_cs_id,omitempty"`
	Amount          float64    `gorm:"column:amount;type:decimal(10,2);not null" json:"amount"`
	SalesCommission float64    `gorm:"column:sales_commission;type:decimal(10,2);default:0" json:"sales_commission"`
	CSCommission    float64    `gorm:"column:cs_commission;type:decimal(10,2);default:0" json:"cs_commission"`
	BillingMonth    time.Time  `gorm:"column:billing_month;type:date;not null" json:"billing_month"`
	Status          string     `gorm:"column:status;type:enum('pending','paid','overdue');default:'pending'" json:"status"`
	Notes           *string    `gorm:"column:notes;type:text" json:"notes,omitempty"`
	FilePath        *string    `gorm:"column:file_path;size:255" json:"file_path,omitempty"`
	FileUploadedAt  *time.Time `gorm:"column:file_uploaded_at" json:"file_uploaded_at,omitempty"`
	PaidAt          *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	PaidBy          *uint      `gorm:"column:paid_by" json:"paid_by,omitempty"`
	CreatedBy       uint       `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	Client        *Client     `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	JobRequest    *JobRequest `gorm:"foreignKey:JobRequestID" json:"job_request,omitempty"`
	AssignedSales *User       `gorm:"foreignKey:AssignedSalesID" json:"assigned_sales,omitempty"`
	AssignedCS    *User       `gorm:"foreignKey:AssignedCSID" json:"assigned_cs,omitempty"`
	PaidByUser    *User       `gorm:"foreignKey:PaidBy" json:"paid_by_user,omitempty"`
	Creator       *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Invoice) TableName() string {
	return "invoices"
}
