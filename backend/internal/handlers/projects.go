package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
)

func (h *Handler) GetProjects(c *gin.Context) {
	var projects []models.Project
	h.DB.Order("id ASC").Find(&projects)
	c.JSON(http.StatusOK, projects)
}

func (h *Handler) GetProjectStaff(c *gin.Context) {
	var project models.Project
	if err := h.DB.First(&project, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Project not found."})
		return
	}

	var members []models.ProjectStaff
	h.DB.Preload("User").
		Where("project_id = ?", project.ID).
		Find(&members)

	c.JSON(http.StatusOK, members)
}

func (h *Handler) AddProjectStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var project models.Project
	if err := h.DB.First(&project, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Project not found."})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "user_id is required."})
		return
	}

	var user models.User
	if err := h.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	var existing models.ProjectStaff
	if err := h.DB.Where("project_id = ? AND user_id = ?", project.ID, req.UserID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"message": "User is already assigned to this project."})
		return
	}

	ps := models.ProjectStaff{
		ProjectID: project.ID,
		UserID:    req.UserID,
	}
	if err := h.DB.Create(&ps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to assign staff."})
		return
	}

	uid := currentUser.ID
	h.Audit.LogCreate(&uid, "ProjectStaff", ps.ID, map[string]interface{}{
		"project_id": project.ID,
		"project":    project.Name,
		"user_id":    req.UserID,
		"user":       user.Name,
	}, c.ClientIP())

	h.DB.Preload("User").Preload("Project").First(&ps, ps.ID)
	c.JSON(http.StatusCreated, ps)
}

func (h *Handler) RemoveProjectStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var project models.Project
	if err := h.DB.First(&project, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Project not found."})
		return
	}

	var ps models.ProjectStaff
	if err := h.DB.Preload("User").
		Where("project_id = ? AND user_id = ?", project.ID, c.Param("userId")).
		First(&ps).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Staff member not found in this project."})
		return
	}

	// Cannot remove yourself
	if ps.UserID == currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot remove yourself from a project."})
		return
	}

	h.DB.Delete(&ps)

	uid := currentUser.ID
	h.Audit.LogDelete(&uid, "ProjectStaff", ps.ID, map[string]interface{}{
		"project_id": project.ID,
		"project":    project.Name,
		"user_id":    ps.UserID,
	}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Staff removed from project."})
}
