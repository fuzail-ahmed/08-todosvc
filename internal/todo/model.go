package todo

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task is the domain and GORM model for todo tasks.
type Task struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	Title       string `gorm:"type:text;not null"`
	Description string `gorm:"type:text"`
	Completed   bool   `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate hook to populate UUID
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}
