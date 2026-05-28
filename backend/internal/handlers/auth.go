package handlers

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/middleware"
	"github.com/trustwired/internal-t/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "These credentials do not match our records."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "These credentials do not match our records."})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"message": "Your account has been deactivated. Contact the Owner."})
		return
	}

	if !user.IsEmailVerified() {
		c.JSON(http.StatusForbidden, gin.H{
			"message":        "Your email address is not verified.",
			"email_verified": false,
		})
		return
	}

	tokenStr, _, err := generateToken(h.DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create token."})
		return
	}

	h.Audit.LogLogin(user.ID, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token":                tokenStr,
		"user":                 user,
		"email_verified":       true,
		"must_change_password": user.MustChangePassword,
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.First(&user, currentUser.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Current password is incorrect."})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	h.DB.Model(&user).Updates(map[string]interface{}{
		"password":             string(hashed),
		"must_change_password": false,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully."})
}

func (h *Handler) Logout(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	tokenID, _ := c.Get("token_id")

	if user != nil {
		h.DB.Delete(&models.PersonalAccessToken{}, tokenID)
		h.Audit.LogLogout(user.ID, c.ClientIP())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully."})
}

func (h *Handler) Me(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var memberships []models.ProjectStaff
	h.DB.Preload("Project").Where("user_id = ?", user.ID).Find(&memberships)

	c.JSON(http.StatusOK, gin.H{
		"id":                 user.ID,
		"name":               user.Name,
		"email":              user.Email,
		"role":               user.Role,
		"is_active":          user.IsActive,
		"email_verified_at":  user.EmailVerifiedAt,
		"created_at":         user.CreatedAt,
		"projects":           memberships,
	})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
		tokenBytes := make([]byte, 32)
		rand.Read(tokenBytes)
		token := hex.EncodeToString(tokenBytes)

		hash := sha256.Sum256([]byte(token))
		hashedToken := hex.EncodeToString(hash[:])
		tokenWithExpiry := fmt.Sprintf("%s|%d", hashedToken, time.Now().Add(60*time.Minute).Unix())

		h.DB.Model(&user).Update("remember_token", tokenWithExpiry)
		h.Email.SendPasswordResetEmail(user.Email, token)
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a password reset link has been sent."})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token                string `json:"token" binding:"required"`
		Email                string `json:"email" binding:"required,email"`
		Password             string `json:"password" binding:"required,min=8"`
		PasswordConfirmation string `json:"password_confirmation" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed.", "errors": err.Error()})
		return
	}

	if req.Password != req.PasswordConfirmation {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Passwords do not match."})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid reset token."})
		return
	}

	if user.RememberToken == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid or expired reset token."})
		return
	}

	hash := sha256.Sum256([]byte(req.Token))
	hashedToken := hex.EncodeToString(hash[:])

	parts := strings.SplitN(*user.RememberToken, "|", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid or expired reset token."})
		return
	}
	storedHash := parts[0]
	expiryTime, _ := strconv.ParseInt(parts[1], 10, 64)

	if storedHash != hashedToken || time.Now().Unix() > expiryTime {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid or expired reset token."})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password."})
		return
	}

	h.DB.Model(&user).Updates(map[string]interface{}{
		"password":       string(hashedPwd),
		"remember_token": nil,
	})

	h.DB.Where("tokenable_type = ? AND tokenable_id = ?", "User", user.ID).Delete(&models.PersonalAccessToken{})

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully."})
}

func (h *Handler) RequestMagicLink(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "A valid email address is required."})
		return
	}

	// Always respond the same way to avoid leaking whether the email exists.
	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err == nil && user.IsActive {
		tokenBytes := make([]byte, 32)
		rand.Read(tokenBytes)
		plainToken := hex.EncodeToString(tokenBytes)

		hash := sha256.Sum256([]byte(plainToken))
		stored := fmt.Sprintf("%s|%d", hex.EncodeToString(hash[:]), time.Now().Add(15*time.Minute).Unix())

		h.DB.Model(&user).Update("remember_token", stored)
		go func() {
			if err := h.Email.SendMagicLinkEmail(user.Email, plainToken); err != nil {
				log.Printf("SendMagicLinkEmail to %s failed: %v", user.Email, err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a sign-in link has been sent."})
}

func (h *Handler) VerifyMagicLink(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired sign-in link."})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"message": "Your account has been deactivated. Contact the Owner."})
		return
	}

	if user.RememberToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired sign-in link."})
		return
	}

	hash := sha256.Sum256([]byte(req.Token))
	hashedToken := hex.EncodeToString(hash[:])

	parts := strings.SplitN(*user.RememberToken, "|", 2)
	if len(parts) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired sign-in link."})
		return
	}
	storedHash := parts[0]
	expiryUnix, _ := strconv.ParseInt(parts[1], 10, 64)

	if storedHash != hashedToken || time.Now().Unix() > expiryUnix {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired sign-in link."})
		return
	}

	// Consume token (one-time use) and verify email if needed
	now := time.Now()
	updates := map[string]interface{}{"remember_token": nil}
	if user.EmailVerifiedAt == nil {
		updates["email_verified_at"] = now
		user.EmailVerifiedAt = &now
	}
	h.DB.Model(&user).Updates(updates)

	tokenStr, _, err := generateToken(h.DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create session."})
		return
	}

	h.Audit.LogLogin(user.ID, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"token": tokenStr,
		"user":  user,
	})
}

func (h *Handler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Validation failed."})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
		if user.EmailVerifiedAt == nil {
			h.Email.SendVerificationEmail(user.ID, user.Email)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent."})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	idStr := c.Param("id")
	hashParam := c.Param("hash")
	expires := c.Query("expires")
	signature := c.Query("signature")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid verification link."})
		return
	}

	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found."})
		return
	}

	if !h.Email.VerifySignedURL(uint(id), user.Email, expires, signature) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Invalid or expired verification link."})
		return
	}

	emailHash := sha1EmailHash(user.Email)
	if hashParam != emailHash {
		c.JSON(http.StatusForbidden, gin.H{"message": "Invalid verification link."})
		return
	}

	if user.EmailVerifiedAt != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Email already verified."})
		return
	}

	now := time.Now()
	h.DB.Model(&user).Update("email_verified_at", now)

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully."})
}

func sha1EmailHash(email string) string {
	h := sha1.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}

func generateToken(db *gorm.DB, userID uint) (string, *models.PersonalAccessToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, err
	}
	plainToken := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	record := &models.PersonalAccessToken{
		TokenableType: "User",
		TokenableID:   userID,
		Name:          "auth_token",
		Token:         tokenHash,
	}

	if err := db.Create(record).Error; err != nil {
		return "", nil, err
	}

	fullToken := fmt.Sprintf("%d|%s", record.ID, plainToken)
	return fullToken, record, nil
}
