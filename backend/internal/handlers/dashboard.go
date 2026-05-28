package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
	"gorm.io/gorm"
)

func (h *Handler) GetDashboard(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	isAdmin := currentUser.HasRole("admin")
	canViewAll := isAdmin || currentUser.HasRole("cs_manager")
	now := time.Now()

	// scopeJobs applies assignment filter for regular users; admins and cs_managers see all
	scopeJobs := func(db *gorm.DB) *gorm.DB {
		if canViewAll {
			return db
		}
		return db.Where("assigned_sales_id = ? OR assigned_cs_id = ?", currentUser.ID, currentUser.ID)
	}

	// Total jobs
	var totalJobs int64
	scopeJobs(h.DB.Model(&models.JobRequest{})).Count(&totalJobs)

	// Active clients scoped to assigned jobs for regular users
	var activeClients int64
	activeClientsQ := h.DB.Model(&models.Client{}).Where("account_status = ?", "active")
	if !canViewAll {
		activeClientsQ = activeClientsQ.Where(
			"id IN (?)",
			h.DB.Model(&models.JobRequest{}).Select("client_id").
				Where("assigned_sales_id = ? OR assigned_cs_id = ?", currentUser.ID, currentUser.ID),
		)
	}
	activeClientsQ.Count(&activeClients)

	// Overdue jobs
	var overdueJobs []models.JobRequest
	scopeJobs(h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS")).
		Where("status != ? AND sla_deadline < ?", "completed", now).
		Find(&overdueJobs)

	// Pending approvals (admin only — non-admins don't approve)
	pendingApprovals := make([]models.Agreement, 0)
	if isAdmin {
		h.DB.Preload("JobRequest").Preload("JobRequest.Client").Preload("Uploader").
			Where("status = ?", "pending_approval").
			Find(&pendingApprovals)
	}

	// Stale jobs
	threeDaysAgo := now.Add(-3 * 24 * time.Hour)
	var staleJobs []models.JobRequest
	scopeJobs(h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS")).
		Where("status != ? AND last_activity_at < ?", "completed", threeDaysAgo).
		Find(&staleJobs)

	// Stuck Stage 2
	var stuckStage2 []models.JobRequest
	scopeJobs(h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").Preload("Tasks")).
		Where("current_stage = ? AND status != ?", 2, "completed").
		Find(&stuckStage2)

	stuckTasks := []models.JobRequest{}
	for _, job := range stuckStage2 {
		for _, task := range job.Tasks {
			if task.Status != "completed" {
				stuckTasks = append(stuckTasks, job)
				break
			}
		}
	}

	// Missing fields
	var missingFields []models.JobRequest
	scopeJobs(h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS")).
		Where("current_stage = 1 AND (customer_pic IS NULL OR monthly_recurring IS NULL OR account_type IS NULL)").
		Find(&missingFields)

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total_jobs":          totalJobs,
			"active_clients":      activeClients,
			"overdue_jobs":        len(overdueJobs),
			"pending_approvals":   len(pendingApprovals),
			"stale_jobs":          len(staleJobs),
			"stuck_stage2_jobs":   len(stuckTasks),
			"missing_fields_jobs": len(missingFields),
		},
		"overdue_jobs":      overdueJobs,
		"pending_approvals": pendingApprovals,
		"stale_jobs":        staleJobs,
		"stuck_tasks":       stuckTasks,
		"missing_fields":    missingFields,
	})
}
