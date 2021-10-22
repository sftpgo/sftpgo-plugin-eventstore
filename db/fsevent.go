package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

// FsEvent defines a filesystem event
type FsEvent struct {
	ID                string `json:"id" gorm:"primaryKey"`
	Timestamp         int64  `json:"timestamp"`
	Action            string `json:"action"`
	Username          string `json:"username"`
	FsPath            string `json:"fs_path"`
	FsTargetPath      string `json:"fs_target_path,omitempty"`
	VirtualPath       string `json:"virtual_path"`
	VirtualTargetPath string `json:"virtual_target_path,omitempty"`
	SSHCmd            string `json:"ssh_cmd,omitempty"`
	FileSize          int64  `json:"file_size,omitempty"`
	Status            int    `json:"status"`
	Protocol          string `json:"protocol"`
	IP                string `json:"ip,omitempty"`
	InstanceID        string `json:"instance_id,omitempty"`
}

// TableName defines the database table name
func (ev *FsEvent) TableName() string {
	return "eventstore_fs_events"
}

// BeforeCreate implements gorm hook
func (ev *FsEvent) BeforeCreate(tx *gorm.DB) (err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		logger.AppLogger.Error("unable to generate uuid", "error", err)
		return err
	}
	ev.ID = uuid.String()
	return
}

// Create persists the object
func (ev *FsEvent) Create(tx *gorm.DB) error {
	return tx.Create(ev).Error
}

func cleanupFsEvents(timestamp time.Time) error {
	logger.AppLogger.Debug("removing fs events", "timestamp", timestamp)
	sess, cancel := getSessionWithTimeout(20 * time.Minute)
	defer cancel()

	sess = sess.Where("timestamp < ?", timestamp.UnixNano()).Delete(&FsEvent{})
	err := sess.Error
	if err == nil {
		logger.AppLogger.Debug("fs events deleted", "num", sess.RowsAffected)
	}
	return err
}
