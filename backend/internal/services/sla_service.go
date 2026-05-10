package services

import (
	"math"
	"time"

	"github.com/trustwired/internal-t/internal/models"
)

type SLAStatus struct {
	Indicator       string     `json:"indicator"`
	SLADeadline     *time.Time `json:"sla_deadline"`
	DaysRemaining   int        `json:"days_remaining"`
	SLAOverdue      bool       `json:"sla_overdue"`
	DaysSinceUpdate float64    `json:"days_since_update"`
	Stale           bool       `json:"stale"`
}

type SLAService struct{}

func NewSLAService() *SLAService {
	return &SLAService{}
}

func (s *SLAService) Calculate(job *models.JobRequest) SLAStatus {
	now := time.Now()
	status := SLAStatus{
		Indicator:   job.Indicator,
		SLADeadline: job.SLADeadline,
	}

	if job.SLADeadline != nil {
		daysRemaining := job.SLADeadline.Sub(now).Hours() / 24
		status.DaysRemaining = int(math.Ceil(daysRemaining))
		status.SLAOverdue = now.After(*job.SLADeadline)
	}

	if job.LastActivityAt != nil {
		daysSince := now.Sub(*job.LastActivityAt).Hours() / 24
		status.DaysSinceUpdate = math.Round(daysSince*10) / 10
		status.Stale = daysSince > 3 && job.Status != "completed"
	}

	return status
}

func (s *SLAService) CalculateIndicator(job *models.JobRequest) string {
	now := time.Now()

	if job.Status == "completed" {
		return "green"
	}

	if job.SLADeadline != nil && now.After(*job.SLADeadline) {
		return "red"
	}

	if job.LastActivityAt != nil {
		daysSince := now.Sub(*job.LastActivityAt).Hours() / 24
		if daysSince > 3 {
			return "red"
		}
	}

	if job.CurrentStage == 1 && job.CustomerPIC == nil && job.MonthlyRecurring == nil && job.AccountType == nil {
		return "grey"
	}

	return "yellow"
}
