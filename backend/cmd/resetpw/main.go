// resetpw resets a user's password by email.
// Usage: go run ./cmd/resetpw/ <email> <new-password>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/trustwired/internal-t/internal/config"
	"github.com/trustwired/internal-t/internal/database"
	"github.com/trustwired/internal-t/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/resetpw/ <email> <new-password>")
		os.Exit(1)
	}

	email := os.Args[1]
	newPassword := os.Args[2]

	if len(newPassword) < 8 {
		log.Fatal("Password must be at least 8 characters")
	}

	_ = godotenv.Load()
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("DB connect failed:", err)
	}

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Fatalf("User not found: %s", email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("bcrypt failed:", err)
	}

	if err := db.Model(&user).Updates(map[string]any{
		"password":             string(hash),
		"remember_token":       nil,
		"must_change_password": false,
	}).Error; err != nil {
		log.Fatal("Failed to update password:", err)
	}

	// Revoke all active sessions
	db.Where("tokenable_id = ? AND tokenable_type = ?", user.ID, "User").Delete(&models.PersonalAccessToken{})

	fmt.Printf("Password reset for %s (%s). All sessions revoked.\n", user.Name, user.Email)
}
