package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	mignationV4ID = "4"
)

type fsEventV4 struct {
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
	SessionID         string `gorm:"size:100"`
	IP                string `gorm:"size:50;index:idx_fs_events_ip"`
	FsProvider        int    `gorm:"size:32;index:idx_fs_events_provider"`
	Bucket            string `gorm:"size:512;index:idx_fs_events_bucket"`
	Endpoint          string `gorm:"size:512;index:idx_fs_events_endpoint"`
	OpenFlags         int    `gorm:"size:32"`
	Role              string `gorm:"size:255;index:idx_fs_events_role"`
	InstanceID        string `gorm:"size:60;index:idx_fs_events_instance_id"`
}

func (ev *fsEventV4) TableName() string {
	return fsEventsTableName
}

type providerEventV4 struct {
	ID         string `gorm:"primaryKey;size:36"`
	Timestamp  int64  `gorm:"size:64;not null;index:idx_provider_events__timestamp"`
	Action     string `gorm:"size:60;not null;index:idx_provider_events_action"`
	Username   string `gorm:"size:255;not null;index:idx_provider_events_username"`
	IP         string `gorm:"size:50;index:idx_provider_events_ip"`
	ObjectType string `gorm:"size:50;index:idx_provider_events_object_type"`
	ObjectName string `gorm:"size:255;index:idx_provider_events_object_name"`
	ObjectData []byte
	Role       string `gorm:"size:255;index:idx_provider_events_role"`
	InstanceID string `gorm:"size:60;index:idx_provider_events_instance_id"`
}

func (ev *providerEventV4) TableName() string {
	return "eventstore_provider_events"
}

func v4Up(tx *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&fsEventV4{},
		&providerEventV4{},
	}
	return tx.AutoMigrate(modelsToMigrate...)
}

func v4Down(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&fsEventV4{}, "Role"); err != nil {
		return err
	}
	return tx.Migrator().DropColumn(&providerEventV4{}, "Role")
}

func getV4Migration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: mignationV4ID,
		Migrate: func(tx *gorm.DB) error {
			return v4Up(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			return v4Down(tx)
		},
	}
}
