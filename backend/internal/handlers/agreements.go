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

func (h *Handler) GetAgreements(c *gin.Context) {
	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	var agreements []models.Agreement
	h.DB.
		Preload("Uploader").
		Preload("Approver").
		Where("job_request_id = ?", job.ID).
		Order("type ASC, version DESC").
		Find(&agreements)

	c.JSON(http.StatusOK, agreements)
}

func (h *Handler) UploadAgreement(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var job models.JobRequest
	if err := h.DB.First(&job, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	isSales := currentUser.HasRole("sales")
	isAdmin := currentUser.HasRole("admin")

	// Sales must be the assigned sales for this job
	if isSales && !isAdmin {
		if job.AssignedSalesID == nil || *job.AssignedSalesID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not the assigned Sales for this job."})
			return
		}
		if job.CurrentStage != 1 {
			c.JSON(http.StatusForbidden, gin.H{"message": "Agreements can only be uploaded during Stage 1."})
			return
		}
	}

	agType := c.PostForm("type")
	if agType != "service_agreement" && agType != "nda" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Type must be 'service_agreement' or 'nda'."})
		return
	}

	notes := c.PostForm("notes")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "File is required."})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".pdf": true, ".doc": true, ".docx": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Only PDF, DOC, and DOCX files are accepted."})
		return
	}

	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "File size must not exceed 10MB."})
		return
	}

	// Get next version
	var maxVersion struct{ Max uint }
	h.DB.Model(&models.Agreement{}).
		Select("COALESCE(MAX(version), 0) as max").
		Where("job_request_id = ? AND type = ?", job.ID, agType).
		Scan(&maxVersion)
	nextVersion := maxVersion.Max + 1

	storagePath := services.GenerateAgreementPath(job.ID, header.Filename)
	if err := h.Storage.Store(file, storagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store file."})
		return
	}

	status := "pending_approval"
	var approvedBy *uint
	var approvedAt *time.Time

	if isAdmin {
		status = "approved"
		uid := currentUser.ID
		approvedBy = &uid
		now := time.Now()
		approvedAt = &now
	}

	agreement := &models.Agreement{
		JobRequestID: job.ID,
		Type:         agType,
		Version:      nextVersion,
		FilePath:     storagePath,
		Status:       status,
		UploadedBy:   currentUser.ID,
		ApprovedBy:   approvedBy,
		ApprovedAt:   approvedAt,
	}
	if notes != "" {
		agreement.Notes = &notes
	}

	if err := h.DB.Create(agreement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save agreement."})
		return
	}

	// Update job status
	if !isAdmin {
		h.DB.Model(&job).Update("status", "pending_to_owner")
	}

	// Update last_activity_at
	now := time.Now()
	h.DB.Model(&job).Update("last_activity_at", now)

	uid := currentUser.ID
	h.Audit.LogCreate(&uid, "Agreement", agreement.ID, map[string]interface{}{
		"type":    agType,
		"version": nextVersion,
		"status":  status,
	}, c.ClientIP())

	h.DB.Preload("Uploader").Preload("Approver").First(agreement, agreement.ID)
	c.JSON(http.StatusCreated, agreement)
}

func (h *Handler) DownloadAgreement(c *gin.Context) {
	var agreement models.Agreement
	if err := h.DB.First(&agreement, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Agreement not found."})
		return
	}

	reader, err := h.Storage.Get(agreement.FilePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found."})
		return
	}
	defer reader.Close()

	ext := filepath.Ext(agreement.FilePath)
	contentType := "application/octet-stream"
	switch strings.ToLower(ext) {
	case ".pdf":
		contentType = "application/pdf"
	case ".doc":
		contentType = "application/msword"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	filename := fmt.Sprintf("agreement-%d-v%d%s", agreement.ID, agreement.Version, ext)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}

func (h *Handler) AddRemarks(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		OwnerRemarks string `json:"owner_remarks" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	var agreement models.Agreement
	if err := h.DB.First(&agreement, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Agreement not found."})
		return
	}

	h.DB.Model(&agreement).Update("owner_remarks", req.OwnerRemarks)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Agreement", agreement.ID, nil, map[string]interface{}{"owner_remarks": req.OwnerRemarks}, c.ClientIP())

	h.DB.Preload("Uploader").Preload("Approver").First(&agreement, agreement.ID)
	c.JSON(http.StatusOK, agreement)
}

func (h *Handler) ApproveAgreement(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var agreement models.Agreement
	if err := h.DB.First(&agreement, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Agreement not found."})
		return
	}

	if agreement.Status != "pending_approval" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Agreement is not pending approval."})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      "approved",
		"approved_by": currentUser.ID,
		"approved_at": now,
	}
	h.DB.Model(&agreement).Updates(updates)

	// Update last_activity_at on job
	h.DB.Model(&models.JobRequest{}).Where("id = ?", agreement.JobRequestID).Update("last_activity_at", now)

	uid := currentUser.ID
	h.Audit.LogApprove(&uid, "Agreement", agreement.ID, c.ClientIP())

	h.DB.Preload("Uploader").Preload("Approver").First(&agreement, agreement.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Agreement approved.", "agreement": agreement})
}

func (h *Handler) RejectAgreement(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		Notes *string `json:"notes"`
	}
	c.ShouldBindJSON(&req)

	var agreement models.Agreement
	if err := h.DB.First(&agreement, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Agreement not found."})
		return
	}

	if agreement.Status != "pending_approval" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Agreement is not pending approval."})
		return
	}

	updates := map[string]interface{}{
		"status": "rejected",
	}
	if req.Notes != nil {
		updates["notes"] = req.Notes
	}
	h.DB.Model(&agreement).Updates(updates)

	// Update job status to client_pending
	now := time.Now()
	h.DB.Model(&models.JobRequest{}).Where("id = ?", agreement.JobRequestID).Updates(map[string]interface{}{
		"status":           "client_pending",
		"last_activity_at": now,
	})

	uid := currentUser.ID
	notesVal := map[string]interface{}{}
	if req.Notes != nil {
		notesVal["notes"] = *req.Notes
	}
	h.Audit.LogReject(&uid, "Agreement", agreement.ID, notesVal, c.ClientIP())

	h.DB.Preload("Uploader").Preload("Approver").First(&agreement, agreement.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Agreement rejected.", "agreement": agreement})
}
