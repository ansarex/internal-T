package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
)

func (h *Handler) GetTasks(c *gin.Context) {
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	var tasks []models.Task
	h.DB.Preload("UpdatedByUser").
		Where("job_request_id = ?", job.ID).
		Order("id ASC").
		Find(&tasks)

	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var task models.Task
	if err := h.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Task not found."})
		return
	}

	var job models.JobRequest
	if err := h.DB.First(&job, task.JobRequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	if job.CurrentStage != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Tasks can only be updated in Stage 2."})
		return
	}

	// Non-admins (cs and cs_manager) must be the assigned CS to update tasks
	if !currentUser.HasRole("admin") {
		if job.AssignedCSID == nil || *job.AssignedCSID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not the assigned CS for this job."})
			return
		}
	}

	var req struct {
		Status  string  `json:"status" binding:"required,oneof=in_progress pending_on_client completed"`
		Remarks *string `json:"remarks"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	oldStatus := task.Status
	now := time.Now()

	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_by": currentUser.ID,
	}
	if req.Remarks != nil {
		updates["remarks"] = req.Remarks
	}
	if req.Status == "completed" {
		updates["completed_at"] = now
	}

	h.DB.Model(&task).Updates(updates)

	// Update last_activity_at on job
	h.DB.Model(&job).Update("last_activity_at", now)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Task", task.ID, map[string]interface{}{"status": oldStatus}, updates, c.ClientIP())

	// Check if all tasks are completed
	if req.Status == "completed" {
		var incompleteTasks int64
		h.DB.Model(&models.Task{}).
			Where("job_request_id = ? AND status != ?", job.ID, "completed").
			Count(&incompleteTasks)

		if incompleteTasks == 0 {
			h.DB.Model(&job).Updates(map[string]interface{}{
				"status":    "completed",
				"indicator": "green",
			})
			h.Audit.LogUpdate(&uid, "JobRequest", job.ID,
				map[string]interface{}{"status": job.Status, "indicator": job.Indicator},
				map[string]interface{}{"status": "completed", "indicator": "green"},
				c.ClientIP())
		}
	}

	h.DB.Preload("UpdatedByUser").First(&task, task.ID)
	c.JSON(http.StatusOK, task)
}
