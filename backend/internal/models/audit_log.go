package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONMap map[string]interface{}

func (j *JSONMap) Scan(v interface{}) error {
	var data []byte
	switch val := v.(type) {
	case []byte:
		data = val
	case string:
		data = []byte(val)
	case nil:
		*j = nil
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return json.Unmarshal(data, j)
}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type AuditLog struct {
	ID            uint      `gorm:"primarykey;column:id" json:"id"`
	UserID        *uint     `gorm:"column:user_id" json:"user_id,omitempty"`
	Action        string    `gorm:"column:action;size:255;not null" json:"action"`
	AuditableType string    `gorm:"column:auditable_type;size:255;not null" json:"auditable_type"`
	AuditableID   uint      `gorm:"column:auditable_id;not null" json:"auditable_id"`
	OldValues     *JSONMap  `gorm:"column:old_values;type:json" json:"old_values,omitempty"`
	NewValues     *JSONMap  `gorm:"column:new_values;type:json" json:"new_values,omitempty"`
	IPAddress     *string   `gorm:"column:ip_address;size:45" json:"ip_address,omitempty"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
