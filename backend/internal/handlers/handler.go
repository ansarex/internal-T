package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/config"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
	"github.com/trustwired/internal-t/internal/services"
	"gorm.io/gorm"
)

// loadJobForSales loads a job and verifies the current sales user is assigned to it.
// Admin bypasses the check. Returns nil if not found or not authorized (response already written).
func (h *Handler) loadJobForSales(c *gin.Context) *models.JobRequest {
	user := middleware.GetCurrentUser(c)
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return nil
	}
	if !user.HasRole("admin") {
		if user.HasRole("sales") && (job.AssignedSalesID == nil || *job.AssignedSalesID != user.ID) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not the assigned Sales for this job."})
			return nil
		}
	}
	return &job
}

// loadJobForCS loads a job and verifies the current CS user is assigned to it.
// Admin bypasses the check.
func (h *Handler) loadJobForCS(c *gin.Context) *models.JobRequest {
	user := middleware.GetCurrentUser(c)
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return nil
	}
	if !user.HasRole("admin") {
		if user.HasRole("cs") && (job.AssignedCSID == nil || *job.AssignedCSID != user.ID) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not the assigned CS for this job."})
			return nil
		}
	}
	return &job
}

type Handler struct {
	DB      *gorm.DB
	Config  *config.Config
	Audit   *services.AuditService
	Email   *services.EmailService
	Storage *services.StorageService
	SLA     *services.SLAService
}

func NewHandler(
	db *gorm.DB,
	cfg *config.Config,
	audit *services.AuditService,
	email *services.EmailService,
	storage *services.StorageService,
	sla *services.SLAService,
) *Handler {
	return &Handler{
		DB:      db,
		Config:  cfg,
		Audit:   audit,
		Email:   email,
		Storage: storage,
		SLA:     sla,
	}
}
