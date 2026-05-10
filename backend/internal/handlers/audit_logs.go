package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/models"
)

func (h *Handler) GetAuditLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage := 50
	offset := (page - 1) * perPage

	var total int64
	h.DB.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	h.DB.Preload("User").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&logs)

	lastPage := int(total) / perPage
	if int(total)%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         logs,
		"total":        total,
		"per_page":     perPage,
		"current_page": page,
		"last_page":    lastPage,
	})
}
