package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
	"github.com/trustwired/internal-t/internal/services"
)

func (h *Handler) GetJobRequests(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var jobs []models.JobRequest
	query := h.DB.
		Preload("Client").
		Preload("AssignedSales").
		Preload("AssignedCS").
		Order("created_at DESC")

	if !currentUser.HasRole("admin") && !currentUser.HasRole("cs_manager") {
		query = query.Where("assigned_sales_id = ? OR assigned_cs_id = ?", currentUser.ID, currentUser.ID)
	}

	if err := query.Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch job requests."})
		return
	}

	type JobWithSLA struct {
		models.JobRequest
		SLA services.SLAStatus `json:"sla"`
	}

	result := make([]JobWithSLA, len(jobs))
	for i, job := range jobs {
		result[i] = JobWithSLA{
			JobRequest: job,
			SLA:        h.SLA.Calculate(&job),
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetJobRequest(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var job models.JobRequest
	if err := h.DB.
		Preload("Client").
		Preload("AssignedSales").
		Preload("AssignedCS").
		Preload("SignedUploader").
		Preload("Stage1Approver").
		Preload("Creator").
		Preload("Tasks").
		Preload("Tasks.UpdatedByUser").
		Preload("Agreements").
		Preload("Agreements.Uploader").
		Preload("Agreements.Approver").
		First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	if !currentUser.HasRole("admin") && !currentUser.HasRole("cs_manager") {
		isAssigned := (job.AssignedSalesID != nil && *job.AssignedSalesID == currentUser.ID) ||
			(job.AssignedCSID != nil && *job.AssignedCSID == currentUser.ID)
		if !isAssigned {
			c.JSON(http.StatusForbidden, gin.H{"message": "You do not have access to this job request."})
			return
		}
	}

	sla := h.SLA.Calculate(&job)

	c.JSON(http.StatusOK, gin.H{
		"job_request": job,
		"sla":         sla,
	})
}

func (h *Handler) GetJobSLA(c *gin.Context) {
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	sla := h.SLA.Calculate(&job)
	c.JSON(http.StatusOK, sla)
}

func (h *Handler) UpdateStage1(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	job := h.loadJobForSales(c)
	if job == nil {
		return
	}

	if job.CurrentStage != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Job request is not in Stage 1."})
		return
	}

	var req struct {
		CustomerPIC         *string  `json:"customer_pic"`
		MonthlyRecurring    *float64 `json:"monthly_recurring"`
		AccountType         *string  `json:"account_type"`
		RecurringStartDate  *string  `json:"recurring_start_date"` // "YYYY-MM-DD"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	oldValues := map[string]interface{}{
		"customer_pic":         job.CustomerPIC,
		"monthly_recurring":    job.MonthlyRecurring,
		"account_type":         job.AccountType,
		"recurring_start_date": job.RecurringStartDate,
	}

	now := time.Now()
	updates := map[string]interface{}{
		"last_activity_at": now,
	}
	if req.CustomerPIC != nil {
		updates["customer_pic"] = req.CustomerPIC
	}
	if req.MonthlyRecurring != nil {
		updates["monthly_recurring"] = req.MonthlyRecurring
	}
	if req.AccountType != nil {
		updates["account_type"] = req.AccountType
	}
	if req.RecurringStartDate != nil {
		if *req.RecurringStartDate == "" {
			updates["recurring_start_date"] = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.RecurringStartDate)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid recurring_start_date. Use YYYY-MM-DD format."})
				return
			}
			updates["recurring_start_date"] = t
		}
	}

	if err := h.DB.Model(&job).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update stage 1."})
		return
	}

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "JobRequest", job.ID, oldValues, updates, c.ClientIP())

	h.DB.First(&job, job.ID)
	c.JSON(http.StatusOK, job)
}

func (h *Handler) UploadSignedCopy(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	job := h.loadJobForSales(c)
	if job == nil {
		return
	}
	// Reload with agreements
	h.DB.Preload("Agreements").First(job, job.ID)

	if job.CurrentStage != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Job request is not in Stage 1."})
		return
	}

	if job.SignedFilePath != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Signed copy already uploaded."})
		return
	}

	// Check both SA and NDA have been uploaded
	hasSA := false
	hasNDA := false
	for _, ag := range job.Agreements {
		if ag.Type == "service_agreement" {
			hasSA = true
		} else if ag.Type == "nda" {
			hasNDA = true
		}
	}

	if !hasSA || !hasNDA {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Both Service Agreement and NDA must be uploaded before uploading signed copy."})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "File is required."})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Only PDF files are accepted."})
		return
	}

	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "File size must not exceed 10MB."})
		return
	}

	storagePath := services.GenerateSignedCopyPath(job.ID, header.Filename)
	if err := h.Storage.Store(file, storagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store file."})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"signed_file_path":   storagePath,
		"signed_uploaded_at": now,
		"signed_uploaded_by": currentUser.ID,
		"stage1_approved_at": now,
		"current_stage":      2,
		"status":             "pending",
		"indicator":          "yellow",
		"last_activity_at":   now,
	}

	h.DB.Model(&job).Updates(updates)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "JobRequest", job.ID, map[string]interface{}{"current_stage": 1}, map[string]interface{}{
		"current_stage": 2,
		"signed_file_path": storagePath,
	}, c.ClientIP())

	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").First(&job, job.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Signed copy uploaded. Stage 2 unlocked.", "job_request": job})
}

// AssignPICs allows support/admin to reassign the Sales and CS PIC on a job request.
func (h *Handler) AssignPICs(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	var req struct {
		AssignedSalesID *uint `json:"assigned_sales_id"`
		AssignedCSID    *uint `json:"assigned_cs_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	oldValues := map[string]interface{}{
		"assigned_sales_id": job.AssignedSalesID,
		"assigned_cs_id":    job.AssignedCSID,
	}
	updates := map[string]interface{}{}

	if req.AssignedSalesID != nil {
		updates["assigned_sales_id"] = req.AssignedSalesID
	}
	if req.AssignedCSID != nil {
		updates["assigned_cs_id"] = req.AssignedCSID
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No fields to update."})
		return
	}

	h.DB.Model(&job).Updates(updates)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "JobRequest", job.ID, oldValues, updates, c.ClientIP())

	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").First(&job, job.ID)
	c.JSON(http.StatusOK, gin.H{"message": "PICs reassigned successfully.", "job_request": job})
}

func (h *Handler) DownloadSignedCopy(c *gin.Context) {
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	if job.SignedFilePath == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No signed copy uploaded yet."})
		return
	}

	reader, err := h.Storage.Get(*job.SignedFilePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found."})
		return
	}
	defer reader.Close()

	filename := fmt.Sprintf("signed-copy-job-%d.pdf", job.ID)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.DataFromReader(http.StatusOK, -1, "application/pdf", reader, nil)
}
