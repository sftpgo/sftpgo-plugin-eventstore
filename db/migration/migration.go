package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
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
		getV2Migration())
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
	m := gormigrate.New(db, options, migrations)
	if err := m.RollbackTo(mignationV1ID); err != nil {
		return err
	}
	if err := v1Down(db); err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%v" CASCADE`, options.TableName)).Error
}
