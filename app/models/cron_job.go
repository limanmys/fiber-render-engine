package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/limanmys/render-engine/internal/database"
	"github.com/limanmys/render-engine/pkg/logger"
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
	if err := database.Connection().Model(&CronJob{}).Where("id = ?", cj.ID).Updates(&CronJob{
		Status:  StatusProcessing,
		Message: "Cronjob processing",
		Output:  "-",
	}).Error; err != nil {
		logger.Sugar().Errorw("error when saving job, %s", err.Error())
	}
}

func (cj *CronJob) UpdateAsFailed(output string) {
	if err := database.Connection().Model(&CronJob{}).Where("id = ?", cj.ID).Updates(&CronJob{
		Status:  StatusFailed,
		Message: "Cronjob failed.",
		Output:  output,
	}).Error; err != nil {
		logger.Sugar().Errorw("error when saving job, %s", err.Error())
	}
}

func (cj *CronJob) UpdateAsDone(output string) {
	if err := database.Connection().Model(&CronJob{}).Where("id = ?", cj.ID).Updates(&CronJob{
		Status:  StatusDone,
		Message: "Cronjob completed successfully. Waiting for next run.",
		Output:  output,
	}).Error; err != nil {
		logger.Sugar().Errorw("error when saving job, %s", err.Error())
	}
}
