package cron

import (
	"log"

	robfigcron "github.com/robfig/cron/v3"
	"github.com/trustwired/internal-t/internal/models"
	"github.com/trustwired/internal-t/internal/services"
	"gorm.io/gorm"
)

func StartCron(db *gorm.DB, slaService *services.SLAService) *robfigcron.Cron {
	c := robfigcron.New()

	c.AddFunc("@hourly", func() {
		RecalculateIndicators(db, slaService)
	})

	c.Start()
	log.Println("Cron jobs started")
	return c
}

func RecalculateIndicators(db *gorm.DB, slaService *services.SLAService) {
	var jobs []models.JobRequest
	if err := db.Where("status != ?", "completed").Find(&jobs).Error; err != nil {
		log.Printf("Cron: failed to fetch job requests: %v", err)
		return
	}

	updated := 0
	for _, job := range jobs {
		newIndicator := slaService.CalculateIndicator(&job)
		if newIndicator != job.Indicator {
			if err := db.Model(&job).Update("indicator", newIndicator).Error; err != nil {
				log.Printf("Cron: failed to update indicator for job %d: %v", job.ID, err)
				continue
			}
			updated++
		}
	}

	log.Printf("Cron: recalculated indicators — %d updated out of %d jobs", updated, len(jobs))
}
