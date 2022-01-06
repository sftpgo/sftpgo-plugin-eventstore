package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	mignationV3ID = "3"
)

type fsEventV3 struct {
	ID                string `gorm:"primaryKey;size:36"`
	Timestamp         int64  `gorm:"size:64;not null;index:idx_fs_events_timestamp"`
	Action            string `gorm:"size:60;not null;index:idx_fs_events_action"`
	Username          string `gorm:"size:255;not null;index:idx_fs_events_username"`
	FsPath            string
	FsTargetPath      string
	VirtualPath       string
	VirtualTargetPath string
	SSHCmd            string `gorm:"size:60;index:idx_fs_events_ssh_cmd"`
	FileSize          int64  `gorm:"size:64"`
	Status            int    `gorm:"size:32;index:idx_fs_events_status"`
	Protocol          string `gorm:"size:30;not null;index:idx_fs_events_protocol"`
	SessionID         string `gorm:"size:100;index:idx_fs_events_session_id"`
	IP                string `gorm:"size:50;index:idx_ip"`
	FsProvider        int    `gorm:"size:32;index:idx_fs_provider"`
	Bucket            string `gorm:"size:512;index:idx_bucket"`
	Endpoint          string `gorm:"size:512;index:idx_endpoint"`
	OpenFlags         int    `gorm:"size:32"`
	InstanceID        string `gorm:"size:60;index:idx_fs_events_instance_id"`
}

func (ev *fsEventV3) TableName() string {
	return fsEventsTableName
}

func v3Up(tx *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&fsEventV3{},
	}
	for _, columnName := range []string{"FsPath", "FsTargetPath", "VirtualPath", "VirtualTargetPath"} {
		if err := tx.Migrator().AlterColumn(&fsEventV3{}, columnName); err != nil {
			return err
		}
	}
	return tx.AutoMigrate(modelsToMigrate...)
}

func v3Down(tx *gorm.DB) error {
	for _, columnName := range []string{"FsProvider", "Bucket", "Endpoint", "OpenFlags"} {
		if err := tx.Migrator().DropColumn(&fsEventV3{}, columnName); err != nil {
			return err
		}
	}
	return nil
}

func getV3Migration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: mignationV3ID,
		Migrate: func(tx *gorm.DB) error {
			return v3Up(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			return v3Down(tx)
		},
	}
}
