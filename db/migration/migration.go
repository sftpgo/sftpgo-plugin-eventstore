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
	"context"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	fsEventsTableName = "eventstore_fs_events"
)

var (
	migrations     []*gormigrate.Migration
	options        *gormigrate.Options
	defaultTimeout = 2 * time.Minute
)

func init() {
	registerMigrations()
	options = gormigrate.DefaultOptions
	options.UseTransaction = true
	options.ValidateUnknownMigrations = true
}

func registerMigrations() {
	migrations = append(migrations,
		getV1Migration(),
		getV2Migration(),
		getV3Migration(),
		getV4Migration(),
		getV5Migration(),
		getV6Migration(),
	)
}

// MigrateDatabase migrates the database to the latest version
func MigrateDatabase(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	db = db.WithContext(ctx)
	m := gormigrate.New(db, options, migrations)
	return m.Migrate()
}

// ResetDatabase removes all the created tables
func ResetDatabase(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	db = db.WithContext(ctx)
	if !db.Migrator().HasTable(options.TableName) {
		fmt.Println("no migration was applied, nothing to do")
		return nil
	}
	m := gormigrate.New(db, options, migrations)
	if err := m.RollbackTo(mignationV1ID); err != nil {
		return err
	}
	if err := v1Down(db); err != nil {
		return err
	}
	return db.Migrator().DropTable(options.TableName)
}
