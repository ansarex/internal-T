package services

import (
	"github.com/trustwired/internal-t/internal/models"
	"gorm.io/gorm"
)

type AuditService struct {
	DB *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{DB: db}
}

func (s *AuditService) Log(userID *uint, action, auditableType string, auditableID uint, oldValues, newValues map[string]interface{}, ipAddress string) {
	log := &models.AuditLog{
		UserID:        userID,
		Action:        action,
		AuditableType: auditableType,
		AuditableID:   auditableID,
		IPAddress:     &ipAddress,
	}

	if oldValues != nil {
		ov := models.JSONMap(oldValues)
		log.OldValues = &ov
	}

	if newValues != nil {
		nv := models.JSONMap(newValues)
		log.NewValues = &nv
	}

	s.DB.Create(log)
}

func (s *AuditService) LogCreate(userID *uint, auditableType string, auditableID uint, newValues map[string]interface{}, ipAddress string) {
	s.Log(userID, "created", auditableType, auditableID, nil, newValues, ipAddress)
}

func (s *AuditService) LogUpdate(userID *uint, auditableType string, auditableID uint, oldValues, newValues map[string]interface{}, ipAddress string) {
	s.Log(userID, "updated", auditableType, auditableID, oldValues, newValues, ipAddress)
}

func (s *AuditService) LogDelete(userID *uint, auditableType string, auditableID uint, oldValues map[string]interface{}, ipAddress string) {
	s.Log(userID, "deleted", auditableType, auditableID, oldValues, nil, ipAddress)
}

func (s *AuditService) LogLogin(userID uint, ipAddress string) {
	uid := userID
	s.Log(&uid, "login", "User", userID, nil, map[string]interface{}{"user_id": userID}, ipAddress)
}

func (s *AuditService) LogLogout(userID uint, ipAddress string) {
	uid := userID
	s.Log(&uid, "logout", "User", userID, nil, nil, ipAddress)
}

func (s *AuditService) LogApprove(userID *uint, auditableType string, auditableID uint, ipAddress string) {
	s.Log(userID, "approved", auditableType, auditableID, nil, nil, ipAddress)
}

func (s *AuditService) LogReject(userID *uint, auditableType string, auditableID uint, notes map[string]interface{}, ipAddress string) {
	s.Log(userID, "rejected", auditableType, auditableID, nil, notes, ipAddress)
}
