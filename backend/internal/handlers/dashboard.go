package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/models"
)

func (h *Handler) GetDashboard(c *gin.Context) {
	now := time.Now()

	// Total jobs
	var totalJobs int64
	h.DB.Model(&models.JobRequest{}).Count(&totalJobs)

	// Active clients
	var activeClients int64
	h.DB.Model(&models.Client{}).Where("account_status = ?", "active").Count(&activeClients)

	// Overdue jobs: status != completed AND sla_deadline < now
	var overdueJobs []models.JobRequest
	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").
		Where("status != ? AND sla_deadline < ?", "completed", now).
		Find(&overdueJobs)

	// Pending approvals: agreements with status = pending_approval
	var pendingApprovals []models.Agreement
	h.DB.Preload("JobRequest").Preload("JobRequest.Client").Preload("Uploader").
		Where("status = ?", "pending_approval").
		Find(&pendingApprovals)

	// Stale jobs: status != completed AND last_activity_at < now - 3 days
	threeDaysAgo := now.Add(-3 * 24 * time.Hour)
	var staleJobs []models.JobRequest
	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").
		Where("status != ? AND last_activity_at < ?", "completed", threeDaysAgo).
		Find(&staleJobs)

	// Stuck Stage 2: current_stage = 2 AND status != completed AND has incomplete tasks
	var stuckStage2 []models.JobRequest
	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").Preload("Tasks").
		Where("current_stage = ? AND status != ?", 2, "completed").
		Find(&stuckStage2)

	// Filter to only those with incomplete tasks
	stuckTasks := []models.JobRequest{}
	for _, job := range stuckStage2 {
		for _, task := range job.Tasks {
			if task.Status != "completed" {
				stuckTasks = append(stuckTasks, job)
				break
			}
		}
	}

	// Missing fields: current_stage = 1 AND any of the fields is NULL
	var missingFields []models.JobRequest
	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").
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
