// Seed creates test accounts for all roles.
// Usage: go run ./cmd/seed/
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/trustwired/internal-t/internal/config"
	"github.com/trustwired/internal-t/internal/database"
	"github.com/trustwired/internal-t/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	Name  string
	Email string
	Roles []string
}

var testAccounts = []seedUser{
	{Name: "Admin Owner", Email: "admin@test.com", Roles: []string{"admin"}},
	{Name: "Support Staff", Email: "support@test.com", Roles: []string{"support"}},
	{Name: "Sales Rep", Email: "sales@test.com", Roles: []string{"sales"}},
	{Name: "CS Agent", Email: "cs@test.com", Roles: []string{"cs"}},
	{Name: "Sales+CS Combo", Email: "salescs@test.com", Roles: []string{"sales", "cs"}},
}

const defaultPassword = "password123"

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("DB connect failed:", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("bcrypt failed:", err)
	}

	now := time.Now()

	fmt.Println("\n=== Seeding test accounts ===")
	for _, u := range testAccounts {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			fmt.Printf("  SKIP  %s (%s) — already exists\n", u.Email, u.Name)
			continue
		}

		user := models.User{
			Name:            u.Name,
			Email:           u.Email,
			Password:        string(hash),
			Role:            models.Roles(u.Roles),
			IsActive:        true,
			EmailVerifiedAt: &now,
		}

		if err := db.Create(&user).Error; err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR %s: %v\n", u.Email, err)
			continue
		}
		fmt.Printf("  OK    %s (%s) roles=%v\n", u.Email, u.Name, u.Roles)
	}

	fmt.Println("\n=== Test credentials ===")
	fmt.Printf("  Password for all accounts: %s\n\n", defaultPassword)
	for _, u := range testAccounts {
		fmt.Printf("  %-12s  %s\n", fmt.Sprintf("[%s]", u.Roles[0]), u.Email)
	}
	fmt.Println()
}
