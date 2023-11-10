package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/limanmys/render-engine/internal/database"
)

type CronJob struct {
	ID        *uuid.UUID `json:"id" gorm:"primary_key,type:uuid"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	ExtensionID *uuid.UUID `json:"extension_id"`
	UserID      *uuid.UUID `json:"user_id"`
	ServerID    *uuid.UUID `json:"server_id"`
	BaseURL     string     `json:"base_url"`

	Payload string `json:"payload"`
	Day     int    `json:"day"`
	Time    string `json:"time"`
	Target  string `json:"target"`

	Message string `json:"message"` // Last run message
	Status  Status `json:"status"`  // Last run status
	Output  string `json:"output"`  // Last run output
}

func (CronJob) TableName() string {
	return "cronjobs"
}

func NewCronJob() *CronJob {
	id := uuid.New()

	return &CronJob{
		ID:      &id,
		Message: "Pending.",
		Status:  StatusPending,
	}
}

func (cj *CronJob) UpdateAsProcessing() {
	cj.Status = StatusProcessing
	cj.Message = "Cronjob processing."
	cj.Output = "-"

	database.Connection().Model(cj).Save(cj)
}

func (cj *CronJob) UpdateAsFailed(message string) {
	cj.Status = StatusFailed
	cj.Message = message

	database.Connection().Model(cj).Save(cj)
}

func (cj *CronJob) UpdateAsDone(output string) {
	cj.Status = StatusDone
	cj.Output = output
	cj.Message = "CronJob completed successfully. Waiting for next run."

	database.Connection().Model(cj).Save(cj)
}
