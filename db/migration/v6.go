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
	mignationV6ID = "6"
)

type logEventV1 struct {
	ID         string `gorm:"primaryKey;size:36"`
	Timestamp  int64  `gorm:"size:64;not null;index:idx_log_events_timestamp"`
	Event      int    `gorm:"size:32;not null;index:idx_log_events_event"`
	Protocol   string `gorm:"size:30;index:idx_log_events_protocol"`
	Username   string `gorm:"size:255;index:idx_log_events_username"`
	IP         string `gorm:"size:50;index:idx_log_events_ip"`
	Message    string
	Role       string `gorm:"size:255;index:idx_log_events_role"`
	InstanceID string `gorm:"size:60;index:idx_log_events_instance_id"`
}

func (ev *logEventV1) TableName() string {
	return "eventstore_log_events"
}

func v6Up(tx *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&logEventV1{},
	}
	return tx.AutoMigrate(modelsToMigrate...)
}

func v6Down(tx *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&logEventV1{},
	}
	return tx.Migrator().DropTable(modelsToMigrate...)
}

func getV6Migration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: mignationV6ID,
		Migrate: func(tx *gorm.DB) error {
			return v6Up(tx)
		},
		Rollback: func(tx *gorm.DB) error {
			return v6Down(tx)
		},
	}
}
