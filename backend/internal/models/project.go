package models

import "time"

type Project struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Slug        string    `gorm:"size:100;not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Project) TableName() string { return "projects" }

type ProjectStaff struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project   *Project  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (ProjectStaff) TableName() string { return "project_staff" }
