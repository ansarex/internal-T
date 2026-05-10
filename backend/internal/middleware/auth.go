package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/models"
	"gorm.io/gorm"
)

func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated."})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		parts := strings.SplitN(tokenStr, "|", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token format."})
			c.Abort()
			return
		}

		tokenID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token."})
			c.Abort()
			return
		}

		plainToken := parts[1]
		hash := sha256.Sum256([]byte(plainToken))
		tokenHash := hex.EncodeToString(hash[:])

		var token models.PersonalAccessToken
		if err := db.Where("id = ? AND token = ?", tokenID, tokenHash).First(&token).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token."})
			c.Abort()
			return
		}

		if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token expired."})
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, token.TokenableID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found."})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"message": "Your account has been deactivated. Contact the Owner."})
			c.Abort()
			return
		}

		// Update last_used_at
		now := time.Now()
		db.Model(&token).Update("last_used_at", now)

		c.Set("user", &user)
		c.Set("token_id", token.ID)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated."})
			c.Abort()
			return
		}

		u := user.(*models.User)
		for _, role := range roles {
			if u.HasRole(role) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to perform this action."})
		c.Abort()
	}
}

func GetCurrentUser(c *gin.Context) *models.User {
	user, _ := c.Get("user")
	if user == nil {
		return nil
	}
	return user.(*models.User)
}
