package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Roles []string

func (r *Roles) Scan(v interface{}) error {
	var data []byte
	switch val := v.(type) {
	case []byte:
		data = val
	case string:
		data = []byte(val)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return json.Unmarshal(data, r)
}

func (r Roles) Value() (driver.Value, error) {
	if r == nil {
		return "[]", nil
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (r Roles) Contains(role string) bool {
	for _, v := range r {
		if v == role {
			return true
		}
	}
	return false
}

type User struct {
	ID                 uint       `gorm:"primarykey;column:id" json:"id"`
	Name               string     `gorm:"column:name;size:255;not null" json:"name"`
	Email              string     `gorm:"column:email;size:255;not null;uniqueIndex" json:"email"`
	Password           string     `gorm:"column:password;size:255;not null" json:"-"`
	Role               Roles      `gorm:"column:role;type:json;not null" json:"role"`
	IsActive           bool       `gorm:"column:is_active;default:true" json:"is_active"`
	MustChangePassword bool       `gorm:"column:must_change_password;default:false" json:"must_change_password"`
	EmailVerifiedAt    *time.Time `gorm:"column:email_verified_at" json:"email_verified_at,omitempty"`
	RememberToken      *string    `gorm:"column:remember_token;size:100" json:"-"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) HasRole(role string) bool {
	return u.Role.Contains(role)
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
