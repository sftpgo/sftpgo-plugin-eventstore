// Copyright (C) 2021-2023 Nicola Murino
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	mignationV2ID = "2"
)

type fsEventV2 struct {
	ID                string `gorm:"primaryKey;size:36"`
	Timestamp         int64  `gorm:"size:64;not null;index:idx_fs_events_timestamp"`
	Action            string `gorm:"size:60;not null;index:idx_fs_events_action"`
	Username          string `gorm:"size:255;not null;index:idx_fs_events_username"`
	FsPath            string `gorm:"size:512"`
	FsTargetPath      string `gorm:"size:512"`
	VirtualPath       string `gorm:"size:512"`
	VirtualTargetPath string `gorm:"size:512"`
	SSHCmd            string `gorm:"size:60;index:idx_fs_events_ssh_cmd"`
	FileSize          int64  `gorm:"size:64"`
	Status            int    `gorm:"size:32;index:idx_fs_events_status"`
	Protocol          string `gorm:"size:30;not null;index:idx_fs_events_protocol"`
	SessionID         string `gorm:"size:100"`
	IP                string `gorm:"size:50;index:idx_fs_events_ip"`
	InstanceID        string `gorm:"size:60;index:idx_fs_events_instance_id"`
}

func (ev *fsEventV2) TableName() string {
	return fsEventsTableName
}

func v2Up(tx *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&fsEventV2{},
	}
	return tx.AutoMigrate(modelsToMigrate...)
}

func v2Down(tx *gorm.DB) error {
	return tx.Migrator().DropColumn(&fsEventV2{}, "SessionID")
}

func getV2Migration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: mignationV2ID,
		Migrate: func(tx *gorm.DB) error {
			return v2Up(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			return v2Down(tx)
		},
	}
}
