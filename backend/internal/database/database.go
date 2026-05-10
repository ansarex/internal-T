package database

import (
	"fmt"
	"log"

	"github.com/trustwired/internal-t/internal/config"
	"github.com/trustwired/internal-t/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBDatabase,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if cfg.AppEnv == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	// Rename legacy 'clients' table to 'customer_crm' before GORM migration.
	// MySQL updates all FK references automatically on RENAME TABLE.
	if err := renameLegacyClients(db); err != nil {
		return fmt.Errorf("rename clients table: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.PersonalAccessToken{},
		&models.Project{},
		&models.ProjectStaff{},
		&models.Client{},
		&models.JobRequest{},
		&models.Agreement{},
		&models.Task{},
		&models.AuditLog{},
		&models.Invoice{},
	); err != nil {
		return err
	}
	return seedProjects(db)
}

// renameLegacyClients renames 'clients' → 'customer_crm' exactly once.
// Handles the recovery case where a previous failed run left an empty customer_crm behind.
func renameLegacyClients(db *gorm.DB) error {
	hasClients := db.Migrator().HasTable("clients")
	hasCRM := db.Migrator().HasTable("customer_crm")

	if !hasClients {
		return nil // Nothing to migrate
	}

	if hasCRM {
		// customer_crm exists — check if it is a stale empty shell from a previous failed run.
		var crmCount, clientsCount int64
		db.Raw("SELECT COUNT(*) FROM `customer_crm`").Scan(&crmCount)
		db.Raw("SELECT COUNT(*) FROM `clients`").Scan(&clientsCount)

		if crmCount == 0 && clientsCount > 0 {
			// Previous migration created the table but never populated it.
			// Drop it so we can proceed with the rename.
			log.Println("Migrating: dropping empty 'customer_crm' left by a previous failed run")
			if err := db.Exec("DROP TABLE `customer_crm`").Error; err != nil {
				return fmt.Errorf("drop stale customer_crm: %w", err)
			}
		} else {
			return nil // Already migrated (or both have data — leave it alone)
		}
	}

	log.Println("Migrating: renaming table 'clients' → 'customer_crm'")

	// Drop every FK on child tables that references 'clients' so GORM can
	// recreate them targeting 'customer_crm' during AutoMigrate.
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	rows, err := sqlDB.Query(`
		SELECT TABLE_NAME, CONSTRAINT_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE REFERENCED_TABLE_SCHEMA = DATABASE()
		  AND REFERENCED_TABLE_NAME   = 'clients'`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tableName, constraintName string
			if rows.Scan(&tableName, &constraintName) == nil {
				if err2 := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`", tableName, constraintName)).Error; err2 != nil {
					log.Printf("  Warning: could not drop FK %s.%s: %v", tableName, constraintName, err2)
				} else {
					log.Printf("  Dropped FK %s.%s → clients", tableName, constraintName)
				}
			}
		}
	}

	if err := db.Exec("RENAME TABLE `clients` TO `customer_crm`").Error; err != nil {
		return err
	}
	log.Println("  Renamed clients → customer_crm ✓")
	return nil
}

func seedProjects(db *gorm.DB) error {
	projects := []models.Project{
		{Name: "Client Tracking", Slug: "crm", Description: "Trustwired internal CRM — client onboarding workflow"},
		{Name: "Investiland", Slug: "investiland", Description: "Investiland project"},
		{Name: "Addhoc TWD", Slug: "addhoc-twd", Description: "Addhoc TWD project"},
	}
	for _, p := range projects {
		db.Where(models.Project{Slug: p.Slug}).FirstOrCreate(&p)
	}
	return nil
}
