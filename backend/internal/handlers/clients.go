package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
)

func (h *Handler) GetClients(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	query := h.DB.Model(&models.Client{}).
		Preload("Creator").
		Preload("JobRequests").
		Preload("JobRequests.AssignedSales").
		Preload("JobRequests.AssignedCS")

	if currentUser.HasRole("sales") && !currentUser.HasRole("support") && !currentUser.HasRole("admin") {
		query = query.Joins("JOIN job_requests jr ON jr.client_id = clients.id").
			Where("jr.assigned_sales_id = ?", currentUser.ID)
	} else if currentUser.HasRole("cs") && !currentUser.HasRole("support") && !currentUser.HasRole("admin") {
		query = query.Joins("JOIN job_requests jr ON jr.client_id = clients.id").
			Where("jr.assigned_cs_id = ?", currentUser.ID)
	}

	var clients []models.Client
	if err := query.Order("clients.created_at DESC").Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch clients."})
		return
	}

	// Calculate total recurring
	var totalRecurring float64
	for _, client := range clients {
		for _, jr := range client.JobRequests {
			if jr.MonthlyRecurring != nil {
				totalRecurring += *jr.MonthlyRecurring
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"clients":          clients,
		"total_recurring":  totalRecurring,
	})
}

func (h *Handler) CreateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		CompanyName     string  `json:"company_name" binding:"required"`
		TodoList        *string `json:"todo_list"`
		AssignedSalesID *uint   `json:"assigned_sales_id"`
		AssignedCSID    *uint   `json:"assigned_cs_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	client := &models.Client{
		CompanyName:   req.CompanyName,
		TodoList:      req.TodoList,
		AccountStatus: "inactive",
		CreatedBy:     currentUser.ID,
	}

	if err := h.DB.Create(client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create client."})
		return
	}

	now := time.Now()
	slaDeadline := now.Add(14 * 24 * time.Hour)

	jobRequest := &models.JobRequest{
		ClientID:        client.ID,
		Status:          "pending",
		CurrentStage:    1,
		Indicator:       "grey",
		AssignedSalesID: req.AssignedSalesID,
		AssignedCSID:    req.AssignedCSID,
		SLAStartedAt:    &now,
		SLADeadline:     &slaDeadline,
		LastActivityAt:  &now,
		CreatedBy:       currentUser.ID,
	}

	if err := h.DB.Create(jobRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create job request."})
		return
	}

	// Create 6 tasks
	for _, taskType := range models.AllTaskTypes {
		task := &models.Task{
			JobRequestID: jobRequest.ID,
			TaskType:     taskType,
			Status:       "pending",
		}
		h.DB.Create(task)

		uid := currentUser.ID
		h.Audit.LogCreate(&uid, "Task", task.ID, map[string]interface{}{
			"task_type":      task.TaskType,
			"status":         task.Status,
			"job_request_id": task.JobRequestID,
		}, c.ClientIP())
	}

	uid := currentUser.ID
	h.Audit.LogCreate(&uid, "Client", client.ID, map[string]interface{}{
		"company_name":   client.CompanyName,
		"account_status": client.AccountStatus,
	}, c.ClientIP())

	h.Audit.LogCreate(&uid, "JobRequest", jobRequest.ID, map[string]interface{}{
		"client_id":    jobRequest.ClientID,
		"status":       jobRequest.Status,
		"current_stage": jobRequest.CurrentStage,
	}, c.ClientIP())

	// Reload with associations
	h.DB.Preload("Creator").Preload("JobRequests").First(client, client.ID)

	c.JSON(http.StatusCreated, client)
}

func (h *Handler) GetClient(c *gin.Context) {
	var client models.Client
	if err := h.DB.
		Preload("Creator").
		Preload("JobRequests").
		Preload("JobRequests.AssignedSales").
		Preload("JobRequests.AssignedCS").
		Preload("JobRequests.Tasks").
		Preload("JobRequests.Tasks.UpdatedByUser").
		Preload("JobRequests.Agreements").
		Preload("JobRequests.Agreements.Uploader").
		Preload("JobRequests.Agreements.Approver").
		First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	c.JSON(http.StatusOK, client)
}

func (h *Handler) UpdateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	var req struct {
		CompanyName     string  `json:"company_name"`
		TodoList        *string `json:"todo_list"`
		AssignedSalesID *uint   `json:"assigned_sales_id"`
		AssignedCSID    *uint   `json:"assigned_cs_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	oldValues := map[string]interface{}{
		"company_name": client.CompanyName,
		"todo_list":    client.TodoList,
	}

	if req.CompanyName != "" {
		client.CompanyName = req.CompanyName
	}
	client.TodoList = req.TodoList

	if err := h.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update client."})
		return
	}

	// Update PICs on the job request if provided
	if req.AssignedSalesID != nil || req.AssignedCSID != nil {
		var jobReq models.JobRequest
		if err := h.DB.Where("client_id = ?", client.ID).First(&jobReq).Error; err == nil {
			jobUpdates := map[string]interface{}{}
			if req.AssignedSalesID != nil {
				jobUpdates["assigned_sales_id"] = req.AssignedSalesID
			}
			if req.AssignedCSID != nil {
				jobUpdates["assigned_cs_id"] = req.AssignedCSID
			}
			h.DB.Model(&jobReq).Updates(jobUpdates)
			uid := currentUser.ID
			h.Audit.LogUpdate(&uid, "JobRequest", jobReq.ID, nil, jobUpdates, c.ClientIP())
		}
	}

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, oldValues, map[string]interface{}{
		"company_name": client.CompanyName,
		"todo_list":    client.TodoList,
	}, c.ClientIP())

	h.DB.Preload("JobRequests").Preload("JobRequests.AssignedSales").Preload("JobRequests.AssignedCS").First(&client, client.ID)
	c.JSON(http.StatusOK, client)
}

// requireAssignedSalesForClient checks that the current sales user is the assigned sales for the client's job.
func (h *Handler) requireAssignedSalesForClient(c *gin.Context, clientID uint) (*models.JobRequest, bool) {
	user := middleware.GetCurrentUser(c)
	if user.HasRole("admin") {
		var job models.JobRequest
		h.DB.Where("client_id = ?", clientID).First(&job)
		return &job, true
	}
	var job models.JobRequest
	if err := h.DB.Where("client_id = ?", clientID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No job request found for this client."})
		return nil, false
	}
	if user.HasRole("sales") && (job.AssignedSalesID == nil || *job.AssignedSalesID != user.ID) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not the assigned Sales for this client."})
		return nil, false
	}
	return &job, true
}

func (h *Handler) ActivateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	jobReq, ok := h.requireAssignedSalesForClient(c, client.ID)
	if !ok {
		return
	}

	if client.AccountStatus == "inactive" {
		if jobReq.Status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Stage 2 must be completed before activating the account."})
			return
		}
	}

	oldStatus := client.AccountStatus
	now := time.Now()

	updates := map[string]interface{}{
		"account_status":               "active",
		"pending_account_status":        nil,
		"pending_status_requested_by":   nil,
		"pending_status_requested_at":   nil,
	}

	h.DB.Model(&client).Updates(updates)
	client.AccountStatus = "active"

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, map[string]interface{}{"account_status": oldStatus}, map[string]interface{}{"account_status": "active"}, c.ClientIP())

	_ = now
	c.JSON(http.StatusOK, gin.H{"message": "Account activated.", "client": client})
}

func (h *Handler) PauseClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	if _, ok := h.requireAssignedSalesForClient(c, client.ID); !ok {
		return
	}

	oldStatus := client.AccountStatus
	var newStatus string
	if client.AccountStatus == "active" {
		newStatus = "paused"
	} else if client.AccountStatus == "paused" {
		newStatus = "active"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot pause/unpause an inactive account."})
		return
	}

	h.DB.Model(&client).Update("account_status", newStatus)
	client.AccountStatus = newStatus

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, map[string]interface{}{"account_status": oldStatus}, map[string]interface{}{"account_status": newStatus}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Account status updated.", "client": client})
}

func (h *Handler) RequestDeactivateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	if _, ok := h.requireAssignedSalesForClient(c, client.ID); !ok {
		return
	}

	now := time.Now()
	inactive := "inactive"
	updates := map[string]interface{}{
		"pending_account_status":      &inactive,
		"pending_status_requested_by": currentUser.ID,
		"pending_status_requested_at": now,
	}

	h.DB.Model(&client).Updates(updates)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, nil, map[string]interface{}{"pending_account_status": "inactive"}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Deactivation request submitted. Awaiting admin approval.", "client": client})
}

func (h *Handler) ApproveDeactivateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	if client.PendingAccountStatus == nil || *client.PendingAccountStatus != "inactive" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No pending deactivation request."})
		return
	}

	updates := map[string]interface{}{
		"account_status":              "inactive",
		"pending_account_status":      nil,
		"pending_status_requested_by": nil,
		"pending_status_requested_at": nil,
	}

	h.DB.Model(&client).Updates(updates)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, map[string]interface{}{"account_status": client.AccountStatus}, map[string]interface{}{"account_status": "inactive"}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Deactivation approved.", "client": client})
}

func (h *Handler) RejectDeactivateClient(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var client models.Client
	if err := h.DB.First(&client, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Client not found."})
		return
	}

	updates := map[string]interface{}{
		"pending_account_status":      nil,
		"pending_status_requested_by": nil,
		"pending_status_requested_at": nil,
	}

	h.DB.Model(&client).Updates(updates)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Client", client.ID, map[string]interface{}{"pending_account_status": "inactive"}, map[string]interface{}{"pending_account_status": nil}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Deactivation request rejected.", "client": client})
}
