package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

// ProviderEvent defines a provider event
type ProviderEvent struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Timestamp  time.Time `json:"timestamp"`
	Action     string    `json:"action"`
	Username   string    `json:"username"`
	IP         string    `json:"ip,omitempty"`
	ObjectType string    `json:"object_type"`
	ObjectName string    `json:"object_name"`
	ObjectData []byte    `json:"object_data"`
	InstanceID string    `json:"instance_id,omitempty"`
}

// TableName defines the database table name
func (ev *ProviderEvent) TableName() string {
	return "eventstore_provider_events"
}

// BeforeCreate implements gorm hook
func (ev *ProviderEvent) BeforeCreate(tx *gorm.DB) (err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logger.AppLogger.Error("unable to generate uuid", "error", err)
		return err
	}
	ev.ID = uuid.String()
	return
}

// Create persists the object
func (ev *ProviderEvent) Create(tx *gorm.DB) error {
	return tx.Create(ev).Error
}

func cleanupProviderEvents(timestamp time.Time) error {
	sess, cancel := getSessionWithTimeout(30 * time.Minute)
	defer cancel()

	logger.AppLogger.Debug("removing provider events", "timestamp", timestamp)
	sess = sess.Where("timestamp < ?", timestamp).Delete(&ProviderEvent{})
	err := sess.Error
	if err == nil {
		logger.AppLogger.Debug("provider events deleted", "num", sess.RowsAffected)
	}
	return err
}
