package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GetStaff(c *gin.Context) {
	roleFilter := c.Query("role")

	query := h.DB.Model(&models.User{})
	if roleFilter != "" {
		query = query.Where("JSON_CONTAINS(role, ?)", `"`+roleFilter+`"`)
	}

	var users []models.User
	if err := query.Order("name ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch staff."})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		Name     string   `json:"name" binding:"required"`
		Email    string   `json:"email" binding:"required,email"`
		Password string   `json:"password" binding:"required,min=8"`
		Role     []string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	// Support cannot create admin roles
	if currentUser.HasRole("support") && !currentUser.HasRole("admin") {
		for _, r := range req.Role {
			if r == "admin" {
				c.JSON(http.StatusForbidden, gin.H{"message": "Support staff cannot create admin accounts."})
				return
			}
		}
	}

	if req.Role == nil {
		req.Role = []string{}
	}

	// Check email uniqueness
	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "The email address is already in use."})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	now := time.Now()
	user := &models.User{
		Name:               req.Name,
		Email:              req.Email,
		Password:           string(hashedPwd),
		Role:               models.Roles(req.Role),
		IsActive:           true,
		MustChangePassword: true,
		EmailVerifiedAt:    &now,
	}

	if err := h.DB.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create staff account."})
		return
	}

	uid := currentUser.ID
	h.Audit.LogCreate(&uid, "User", user.ID, map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}, c.ClientIP())

	go h.Email.SendWelcomeEmail(user.Name, user.Email, req.Password)

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) UpdateStaffRole(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		Role []string `json:"role" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Staff not found."})
		return
	}

	oldRole := user.Role
	user.Role = models.Roles(req.Role)

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update role."})
		return
	}

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "User", user.ID, map[string]interface{}{
		"role": oldRole,
	}, map[string]interface{}{
		"role": user.Role,
	}, c.ClientIP())

	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeactivateStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var target models.User
	if err := h.DB.First(&target, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Staff not found."})
		return
	}

	if target.ID == currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot deactivate your own account."})
		return
	}

	h.DB.Model(&target).Update("is_active", false)

	// Revoke all tokens
	h.DB.Where("tokenable_type = ? AND tokenable_id = ?", "User", target.ID).Delete(&models.PersonalAccessToken{})

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "User", target.ID, map[string]interface{}{"is_active": true}, map[string]interface{}{"is_active": false}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Staff account deactivated.", "user": target})
}

func (h *Handler) ActivateStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var target models.User
	if err := h.DB.First(&target, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Staff not found."})
		return
	}

	h.DB.Model(&target).Update("is_active", true)

	uid := currentUser.ID
	h.Audit.LogUpdate(&uid, "User", target.ID, map[string]interface{}{"is_active": false}, map[string]interface{}{"is_active": true}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Staff account activated.", "user": target})
}

func (h *Handler) DeleteStaff(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var target models.User
	if err := h.DB.First(&target, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Staff not found."})
		return
	}

	if target.ID == currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete your own account."})
		return
	}

	// Revoke all tokens first
	h.DB.Where("tokenable_type = ? AND tokenable_id = ?", "User", target.ID).Delete(&models.PersonalAccessToken{})

	// Null out FK references via DB constraints (ON DELETE SET NULL handles this)
	h.DB.Delete(&target)

	uid := currentUser.ID
	h.Audit.LogDelete(&uid, "User", target.ID, map[string]interface{}{
		"name":  target.Name,
		"email": target.Email,
	}, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "Staff account deleted."})
}
