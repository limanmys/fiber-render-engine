package models

import (
	"time"

	gormjsonb "github.com/dariubs/gorm-jsonb"
	"github.com/google/uuid"
	"github.com/limanmys/render-engine/internal/database"
	"gorm.io/gorm"
)

type Operation string
type Status string

const (
	OperationCreate  Operation = "create"
	OperationUpdate  Operation = "update"
	OperationInstall Operation = "install"

	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

// Queue structure of Queue object
type Queue struct {
	ID        string          `json:"id"`
	Type      Operation       `json:"type"`
	Status    Status          `json:"status"`
	Data      gormjsonb.JSONB `json:"data" gorm:"type:jsonb;index,type:gin"`
	Error     string          `json:"error"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (Queue) TableName() string {
	return "queue"
}

// Fill ID of Queue object with UUID beforeCreate
func (q *Queue) BeforeCreate(tx *gorm.DB) (err error) {
	q.ID = uuid.New().String()
	q.Status = StatusPending
	q.CreatedAt = time.Now()
	q.UpdatedAt = time.Now()
	return
}

func (q *Queue) BeforeUpdate(tx *gorm.DB) (err error) {
	q.UpdatedAt = time.Now()
	return
}

func (q *Queue) UpdateStatus(status Status) {
	q.Status = status
	database.Connection().Model(q).Save(q)
}

func (q *Queue) UpdateError(err string) {
	q.Error = err
	q.Status = StatusFailed
	database.Connection().Model(q).Save(q)
}
