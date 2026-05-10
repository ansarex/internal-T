package models

import "time"

type Agreement struct {
	ID           uint       `gorm:"primarykey;column:id" json:"id"`
	JobRequestID uint       `gorm:"column:job_request_id;not null" json:"job_request_id"`
	Type         string     `gorm:"column:type;type:enum('service_agreement','nda');not null" json:"type"`
	Version      uint       `gorm:"column:version;default:1" json:"version"`
	FilePath     string     `gorm:"column:file_path;size:255;not null" json:"file_path"`
	Status       string     `gorm:"column:status;type:enum('draft','pending_approval','approved','rejected');default:'draft'" json:"status"`
	UploadedBy   uint       `gorm:"column:uploaded_by;not null" json:"uploaded_by"`
	ApprovedBy   *uint      `gorm:"column:approved_by" json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	Notes        *string    `gorm:"column:notes;type:text" json:"notes,omitempty"`
	OwnerRemarks *string    `gorm:"column:owner_remarks;type:text" json:"owner_remarks,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	JobRequest *JobRequest `gorm:"foreignKey:JobRequestID" json:"job_request,omitempty"`
	Uploader   *User       `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	Approver   *User       `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (Agreement) TableName() string {
	return "agreements"
}
