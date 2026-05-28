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
	"gorm.io/gorm"
)

func (h *Handler) GetReceipts(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	query := h.DB.Model(&models.Receipt{}).
		Preload("Client").
		Preload("AssignedSales").
		Preload("AssignedCS").
		Preload("PaidByUser")

	if monthStr := c.Query("month"); monthStr != "" {
		month, err := time.Parse("2006-01-02", monthStr)
		if err == nil {
			startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := startOfMonth.AddDate(0, 1, 0)
			query = query.Where("billing_month >= ? AND billing_month < ?", startOfMonth, endOfMonth)
		}
	}

	if statusFilter := c.Query("status"); statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if !currentUser.HasRole("admin") && currentUser.HasRole("sales") {
		query = query.Where("assigned_sales_id = ?", currentUser.ID)
	}

	var receipts []models.Receipt
	if err := query.Order("created_at DESC").Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch receipts."})
		return
	}

	c.JSON(http.StatusOK, receipts)
}

func (h *Handler) GetActiveClientsForReceipt(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01-02"))
	month, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		month = time.Now()
	}
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := h.DB.Model(&models.Client{}).
		Preload("JobRequests", "status = ?", "completed").
		Preload("JobRequests.AssignedSales").
		Preload("JobRequests.AssignedCS").
		Where("account_status = ?", "active")

	if !currentUser.HasRole("admin") && currentUser.HasRole("sales") {
		query = query.Joins("JOIN job_requests jr ON jr.client_id = customer_crm.id AND jr.assigned_sales_id = ?", currentUser.ID)
	}

	var clients []models.Client
	query.Find(&clients)

	now := time.Now()
	isOverdueMissing := now.Day() > 5 && now.Year() == billingMonth.Year() && now.Month() == billingMonth.Month()

	type ClientReceiptStatus struct {
		ClientID         uint          `json:"client_id"`
		CompanyName      string        `json:"company_name"`
		JobRequestID     *uint         `json:"job_request_id"`
		MonthlyRecurring *float64      `json:"monthly_recurring"`
		AssignedSales    *models.User  `json:"assigned_sales"`
		AssignedCS       *models.User  `json:"assigned_cs"`
		Receipt          *models.Receipt `json:"receipt"`
		Receipted        bool          `json:"receipted"`
		OverdueMissing   bool          `json:"overdue_missing"`
	}

	result := []ClientReceiptStatus{}
	for _, client := range clients {
		status := ClientReceiptStatus{
			ClientID:    client.ID,
			CompanyName: client.CompanyName,
		}

		if len(client.JobRequests) > 0 {
			jr := client.JobRequests[0]
			status.JobRequestID = &jr.ID
			status.MonthlyRecurring = jr.MonthlyRecurring
			status.AssignedSales = jr.AssignedSales
			status.AssignedCS = jr.AssignedCS
		}

		var receipt models.Receipt
		err := h.DB.Where("client_id = ? AND billing_month = ?", client.ID, billingMonth).
			Preload("AssignedSales").Preload("AssignedCS").Preload("PaidByUser").
			First(&receipt).Error
		if err == nil {
			status.Receipt = &receipt
			status.Receipted = true
		} else {
			status.OverdueMissing = isOverdueMissing
		}

		result = append(result, status)
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetAdminReceiptOverview(c *gin.Context) {
	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01-02"))
	month, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		month = time.Now()
	}
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	var clients []models.Client
	h.DB.Model(&models.Client{}).
		Preload("JobRequests", func(db *gorm.DB) *gorm.DB {
			return db.Preload("AssignedSales").Preload("AssignedCS")
		}).
		Where("account_status = ?", "active").
		Find(&clients)

	now := time.Now()
	isOverdueMissing := now.Day() > 5 && now.Year() == billingMonth.Year() && now.Month() == billingMonth.Month()

	type ClientRow struct {
		ClientID         uint             `json:"client_id"`
		CompanyName      string           `json:"company_name"`
		JobRequestID     *uint            `json:"job_request_id"`
		MonthlyRecurring *float64         `json:"monthly_recurring"`
		AssignedSales    *models.User     `json:"assigned_sales"`
		AssignedCS       *models.User     `json:"assigned_cs"`
		Receipt          *models.Receipt  `json:"receipt"`
		Receipted        bool             `json:"receipted"`
		OverdueMissing   bool             `json:"overdue_missing"`
	}

	rows := []ClientRow{}
	var totalAmount float64
	receipted, missing := 0, 0

	for _, client := range clients {
		row := ClientRow{
			ClientID:    client.ID,
			CompanyName: client.CompanyName,
		}
		if len(client.JobRequests) > 0 {
			jr := client.JobRequests[0]
			row.JobRequestID = &jr.ID
			row.MonthlyRecurring = jr.MonthlyRecurring
			row.AssignedSales = jr.AssignedSales
			row.AssignedCS = jr.AssignedCS
		}

		var receipt models.Receipt
		err := h.DB.Where("client_id = ? AND billing_month = ?", client.ID, billingMonth).
			Preload("AssignedSales").Preload("AssignedCS").Preload("PaidByUser").
			First(&receipt).Error
		if err == nil {
			row.Receipt = &receipt
			row.Receipted = true
			totalAmount += receipt.Amount
			receipted++
		} else {
			row.OverdueMissing = isOverdueMissing
			missing++
		}
		rows = append(rows, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": rows,
		"summary": gin.H{
			"total_clients": len(clients),
			"receipted":     receipted,
			"missing":       missing,
			"total_amount":  totalAmount,
		},
	})
}

func (h *Handler) GetCommissions(c *gin.Context) {
	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01-02"))
	month, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		month = time.Now()
	}
	billingMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

	type CommissionRow struct {
		StaffID      uint    `json:"staff_id"`
		Name         string  `json:"name"`
		Role         string  `json:"role"`
		Commission   float64 `json:"commission"`
		ReceiptCount int     `json:"receipt_count"`
	}

	var receipts []models.Receipt
	h.DB.
		Preload("AssignedSales").
		Preload("AssignedCS").
		Where("billing_month = ? AND status = ?", billingMonth, "paid").
		Find(&receipts)

	commissions := map[uint]*CommissionRow{}
	for _, r := range receipts {
		if r.AssignedSalesID != nil && r.AssignedSales != nil {
			row, ok := commissions[*r.AssignedSalesID]
			if !ok {
				row = &CommissionRow{StaffID: *r.AssignedSalesID, Name: r.AssignedSales.Name, Role: "sales"}
				commissions[*r.AssignedSalesID] = row
			}
			row.Commission += r.SalesCommission
			row.ReceiptCount++
		}
		if r.AssignedCSID != nil && r.AssignedCS != nil {
			row, ok := commissions[*r.AssignedCSID]
			if !ok {
				row = &CommissionRow{StaffID: *r.AssignedCSID, Name: r.AssignedCS.Name, Role: "cs"}
				commissions[*r.AssignedCSID] = row
			}
			row.Commission += r.CSCommission
			row.ReceiptCount++
		}
	}

	rows := []*CommissionRow{}
	for _, r := range commissions {
		rows = append(rows, r)
	}

	c.JSON(http.StatusOK, rows)
}

func (h *Handler) CreateReceipt(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var clientID uint
	var jobRequestID uint
	var amount float64
	var billingMonthStr string
	var notes string

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		var req struct {
			ClientID     uint    `json:"client_id" binding:"required"`
			JobRequestID uint    `json:"job_request_id" binding:"required"`
			Amount       float64 `json:"amount" binding:"required"`
			BillingMonth string  `json:"billing_month" binding:"required"`
			Notes        string  `json:"notes"`
		}
		if err2 := c.ShouldBindJSON(&req); err2 != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
			return
		}
		clientID = req.ClientID
		jobRequestID = req.JobRequestID
		amount = req.Amount
		billingMonthStr = req.BillingMonth
		notes = req.Notes
	} else {
		fmt.Sscanf(c.PostForm("client_id"), "%d", &clientID)
		fmt.Sscanf(c.PostForm("job_request_id"), "%d", &jobRequestID)
		fmt.Sscanf(c.PostForm("amount"), "%f", &amount)
		billingMonthStr = c.PostForm("billing_month")
		notes = c.PostForm("notes")
	}

	billingMonth, err := time.Parse("2006-01-02", billingMonthStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid billing_month format. Use YYYY-MM-DD."})
		return
	}
	billingMonth = time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, time.UTC)

	var jobReq models.JobRequest
	if err := h.DB.First(&jobReq, jobRequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Job request not found."})
		return
	}

	receiptNumber := generateReceiptNumber(h.DB, billingMonth)

	receipt := &models.Receipt{
		ReceiptNumber:   receiptNumber,
		ClientID:        clientID,
		JobRequestID:    jobRequestID,
		AssignedSalesID: jobReq.AssignedSalesID,
		AssignedCSID:    jobReq.AssignedCSID,
		Amount:          amount,
		SalesCommission: amount * 0.10,
		CSCommission:    amount * 0.10,
		BillingMonth:    billingMonth,
		Status:          "pending",
		CreatedBy:       currentUser.ID,
	}
	if notes != "" {
		receipt.Notes = &notes
	}

	file, header, fileErr := c.Request.FormFile("file")
	if fileErr == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".pdf" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Only PDF files are accepted."})
			return
		}
		storagePath := services.GenerateReceiptPath(billingMonth, header.Filename)
		if err := h.Storage.Store(file, storagePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store file."})
			return
		}
		now := time.Now()
		receipt.FilePath = &storagePath
		receipt.FileUploadedAt = &now
	}

	if err := h.DB.Create(receipt).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "unique") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "A receipt already exists for this client and month."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create receipt."})
		return
	}

	uid := currentUser.ID
	h.Audit.LogCreate(&uid, "Receipt", receipt.ID, map[string]any{
		"receipt_number": receipt.ReceiptNumber,
		"amount":         receipt.Amount,
		"client_id":      receipt.ClientID,
	}, c.ClientIP())

	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").First(receipt, receipt.ID)
	c.JSON(http.StatusCreated, receipt)
}

func (h *Handler) UploadReceiptFile(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var receipt models.Receipt
	if err := h.DB.First(&receipt, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Receipt not found."})
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

	storagePath := services.GenerateReceiptPath(receipt.BillingMonth, header.Filename)
	if err := h.Storage.Store(file, storagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to store file."})
		return
	}

	now := time.Now()
	h.DB.Model(&receipt).Updates(map[string]any{
		"file_path":        storagePath,
		"file_uploaded_at": now,
	})

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Receipt", receipt.ID, nil, map[string]any{"file_path": storagePath}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded.", "receipt": receipt})
}

func (h *Handler) DownloadReceipt(c *gin.Context) {
	var receipt models.Receipt
	if err := h.DB.First(&receipt, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Receipt not found."})
		return
	}

	if receipt.FilePath == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No file uploaded for this receipt."})
		return
	}

	reader, err := h.Storage.Get(*receipt.FilePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found."})
		return
	}
	defer reader.Close()

	filename := fmt.Sprintf("%s.pdf", receipt.ReceiptNumber)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.DataFromReader(http.StatusOK, -1, "application/pdf", reader, nil)
}

func (h *Handler) PayReceipt(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var receipt models.Receipt
	if err := h.DB.First(&receipt, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Receipt not found."})
		return
	}

	if receipt.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Receipt is already paid."})
		return
	}

	now := time.Now()
	h.DB.Model(&receipt).Updates(map[string]any{
		"status":  "paid",
		"paid_at": now,
		"paid_by": currentUser.ID,
	})

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Receipt", receipt.ID,
		map[string]any{"status": receipt.Status},
		map[string]any{"status": "paid", "paid_by": currentUser.ID},
		c.ClientIP())

	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").Preload("PaidByUser").First(&receipt, receipt.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Receipt marked as paid.", "receipt": receipt})
}

func (h *Handler) UpdateReceipt(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var receipt models.Receipt
	if err := h.DB.First(&receipt, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Receipt not found."})
		return
	}

	var req struct {
		Status string   `json:"status"`
		Notes  *string  `json:"notes"`
		Amount *float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	if req.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Use the /pay endpoint to mark receipts as paid."})
		return
	}

	updates := map[string]any{}
	oldValues := map[string]any{}

	if req.Status != "" && (req.Status == "pending" || req.Status == "overdue") {
		oldValues["status"] = receipt.Status
		updates["status"] = req.Status
	}
	if req.Notes != nil {
		oldValues["notes"] = receipt.Notes
		updates["notes"] = req.Notes
	}
	if req.Amount != nil {
		oldValues["amount"] = receipt.Amount
		updates["amount"] = *req.Amount
		updates["sales_commission"] = *req.Amount * 0.10
		updates["cs_commission"] = *req.Amount * 0.10
	}

	if len(updates) > 0 {
		h.DB.Model(&receipt).Updates(updates)
	}

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "Receipt", receipt.ID, oldValues, updates, c.ClientIP())

	h.DB.Preload("Client").Preload("AssignedSales").Preload("AssignedCS").First(&receipt, receipt.ID)
	c.JSON(http.StatusOK, receipt)
}

func generateReceiptNumber(db *gorm.DB, billingMonth time.Time) string {
	prefix := fmt.Sprintf("RCP-%s-", billingMonth.Format("200601"))
	var count int64
	db.Model(&models.Receipt{}).
		Where("receipt_number LIKE ?", prefix+"%").
		Count(&count)
	return fmt.Sprintf("%s%04d", prefix, count+1)
}
