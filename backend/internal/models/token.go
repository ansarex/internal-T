package models

import "time"

type PersonalAccessToken struct {
	ID            uint       `gorm:"primarykey;column:id" json:"id"`
	TokenableType string     `gorm:"column:tokenable_type;size:255;not null;index:idx_tokenable" json:"tokenable_type"`
	TokenableID   uint       `gorm:"column:tokenable_id;not null;index:idx_tokenable" json:"tokenable_id"`
	Name          string     `gorm:"column:name;size:255;not null" json:"name"`
	Token         string     `gorm:"column:token;size:64;not null;uniqueIndex" json:"-"`
	Abilities     *string    `gorm:"column:abilities" json:"abilities,omitempty"`
	LastUsedAt    *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PersonalAccessToken) TableName() string {
	return "personal_access_tokens"
}
