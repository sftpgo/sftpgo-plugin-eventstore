// Copyright (C) 2026 Nicola Murino
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

package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

const (
	schemaVersion    = 1
	migrationTimeout = 2 * time.Minute
)

// MigrateDatabase applies all pending database migrations
func MigrateDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	conn, err := dbHandle.Conn(ctx)
	if err != nil {
		return fmt.Errorf("unable to get connection from pool: %w", err)
	}
	defer conn.Close()

	if err := acquireLock(conn); err != nil {
		return err
	}
	defer releaseLock(conn)

	// Check for gormigrate legacy table
	if tableExists(ctx, conn, "migrations") {
		var fromV6 bool
		switch {
		case hasGormigrateMigration(ctx, conn, "7"):
			logger.AppLogger.Info("found gormigrate legacy table at version 7, migrating to eventstore_schema_version")
		case hasGormigrateMigration(ctx, conn, "6"):
			logger.AppLogger.Info("found gormigrate legacy table at version 6, migrating to eventstore_schema_version")
			fromV6 = true
		default:
			return errors.New("unsupported gormigrate version, please upgrade to v1.0.25 first")
		}
		if err := migrateFromGormigrate(ctx, conn, fromV6); err != nil {
			return fmt.Errorf("unable to migrate from gormigrate: %w", err)
		}
		logger.AppLogger.Info("migration from gormigrate completed successfully")
	}

	if tableExists(ctx, conn, "eventstore_schema_version") {
		hasVersion, err := hasSchemaVersion(ctx, conn)
		if err != nil {
			return fmt.Errorf("unable to read the schema version: %w", err)
		}
		if hasVersion {
			return applyMigrations(ctx, conn)
		}
		logger.AppLogger.Warn("found an interrupted initial install, completing it")
	} else {
		logger.AppLogger.Info("creating initial database schema")
	}

	if err := initializeDatabase(ctx, conn); err != nil {
		return err
	}
	return applyMigrations(ctx, conn)
}

func hasSchemaVersion(ctx context.Context, conn *sql.Conn) (bool, error) {
	var exists bool
	err := conn.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM eventstore_schema_version)").Scan(&exists)
	return exists, err
}

func applyMigrations(ctx context.Context, conn *sql.Conn) error {
	version, err := getSchemaVersion(ctx, conn)
	if err != nil {
		return fmt.Errorf("unable to get schema version: %w", err)
	}
	switch {
	case version == schemaVersion:
		logger.AppLogger.Debug("database schema is up to date", "version", version)
		return nil
	case version > schemaVersion:
		logger.AppLogger.Warn("database schema version is newer than the supported one",
			"current", version, "supported", schemaVersion)
		return nil
	default:
		return fmt.Errorf("unsupported database schema version %d, expected %d", version, schemaVersion)
	}
}

// ResetDatabase drops all tables
func ResetDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	conn, err := dbHandle.Conn(ctx)
	if err != nil {
		return fmt.Errorf("unable to get connection from pool: %w", err)
	}
	defer conn.Close()

	if err := acquireLock(conn); err != nil {
		return err
	}
	defer releaseLock(conn)

	for _, table := range []string{
		"eventstore_fs_events",
		"eventstore_provider_events",
		"eventstore_log_events",
		"eventstore_schema_version",
		"migrations",
	} {
		if err := dropTable(ctx, conn, table); err != nil {
			return err
		}
	}
	return nil
}

func acquireLock(conn *sql.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	switch driverName {
	case driverNamePostgreSQL:
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(201,1)`)
		if err != nil {
			return fmt.Errorf("unable to get advisory lock: %w", err)
		}
		logger.AppLogger.Info("acquired database lock")
	case driverNameMySQL:
		var lockResult sql.NullInt64
		err := conn.QueryRowContext(ctx, `SELECT GET_LOCK('sftpgo_events.migration',30)`).Scan(&lockResult)
		if err != nil {
			return fmt.Errorf("unable to get lock: %w", err)
		}
		if !lockResult.Valid {
			return errors.New("unable to get lock: null value returned")
		}
		if lockResult.Int64 != 1 {
			return fmt.Errorf("unable to get lock, result: %d", lockResult.Int64)
		}
		logger.AppLogger.Info("acquired database lock")
	}
	return nil
}

func releaseLock(conn *sql.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	switch driverName {
	case driverNamePostgreSQL:
		_, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock(201,1)`)
		if err != nil {
			logger.AppLogger.Warn("unable to release lock", "error", err)
		} else {
			logger.AppLogger.Info("released database lock")
		}
	case driverNameMySQL:
		_, err := conn.ExecContext(ctx, `SELECT RELEASE_LOCK('sftpgo_events.migration')`)
		if err != nil {
			logger.AppLogger.Warn("unable to release lock", "error", err)
		} else {
			logger.AppLogger.Info("released database lock")
		}
	}
}

func tableExists(ctx context.Context, conn *sql.Conn, table string) bool {
	var exists bool
	switch driverName {
	case driverNamePostgreSQL:
		err := conn.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = $1 AND table_schema = current_schema())",
			table).Scan(&exists)
		return err == nil && exists
	case driverNameMySQL:
		err := conn.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE())", table).Scan(&exists)
		return err == nil && exists
	}
	return false
}

func hasGormigrateMigration(ctx context.Context, conn *sql.Conn, migrationID string) bool {
	var id string
	var err error
	switch driverName {
	case driverNamePostgreSQL:
		err = conn.QueryRowContext(ctx, "SELECT id FROM migrations WHERE id = $1", migrationID).Scan(&id)
	default:
		err = conn.QueryRowContext(ctx, "SELECT id FROM migrations WHERE id = ?", migrationID).Scan(&id)
	}
	return err == nil && id == migrationID
}

func getSchemaVersion(ctx context.Context, conn *sql.Conn) (int, error) {
	var version sql.NullInt64
	err := conn.QueryRowContext(ctx, "SELECT MAX(version) FROM eventstore_schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, errors.New("eventstore_schema_version table is empty")
	}
	return int(version.Int64), nil
}

func createSchemaVersionTable(ctx context.Context, conn *sql.Conn, version int) error {
	var createSQL string
	switch driverName {
	case driverNameMySQL:
		createSQL = "CREATE TABLE eventstore_schema_version (id int AUTO_INCREMENT NOT NULL PRIMARY KEY, version int NOT NULL)"
	case driverNamePostgreSQL:
		createSQL = `CREATE TABLE eventstore_schema_version (id integer NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY, version int NOT NULL)`
	}
	if _, err := conn.ExecContext(ctx, createSQL); err != nil {
		return err
	}
	var insertSQL string
	switch driverName {
	case driverNamePostgreSQL:
		insertSQL = "INSERT INTO eventstore_schema_version (version) VALUES ($1)"
	default:
		insertSQL = "INSERT INTO eventstore_schema_version (version) VALUES (?)"
	}
	_, err := conn.ExecContext(ctx, insertSQL, version)
	return err
}

func dropTable(ctx context.Context, conn *sql.Conn, table string) error {
	var q string
	switch driverName {
	case driverNameMySQL:
		q = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)
	default:
		q = fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, table)
	}
	_, err := conn.ExecContext(ctx, q)
	return err
}

func initializeDatabase(ctx context.Context, conn *sql.Conn) error {
	var initialSQL string
	switch driverName {
	case driverNameMySQL:
		initialSQL = mysqlInitialSQL
	case driverNamePostgreSQL:
		initialSQL = pgsqlInitialSQL
	default:
		return fmt.Errorf("unsupported driver: %q", driverName)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}

	for _, q := range strings.Split(initialSQL, ";") {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, q); err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to execute query %q: %w", q, err)
		}
	}

	return tx.Commit()
}

func migrateFromGormigrate(ctx context.Context, conn *sql.Conn, fromV6 bool) error {
	var createSQL, dropSQL, v6ToV7SQL string
	switch driverName {
	case driverNameMySQL:
		createSQL = "CREATE TABLE IF NOT EXISTS eventstore_schema_version (id int AUTO_INCREMENT NOT NULL PRIMARY KEY, version int NOT NULL)"
		dropSQL = "DROP TABLE IF EXISTS `migrations`"
		v6ToV7SQL = mysqlGormigrateV6ToV7
	case driverNamePostgreSQL:
		createSQL = `CREATE TABLE IF NOT EXISTS eventstore_schema_version (id integer NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY, version int NOT NULL)`
		dropSQL = `DROP TABLE IF EXISTS "migrations"`
		v6ToV7SQL = pgsqlGormigrateV6ToV7
	default:
		return fmt.Errorf("unsupported driver: %q", driverName)
	}
	const insertSQL = "INSERT INTO eventstore_schema_version (version) SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM eventstore_schema_version)"

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}

	if fromV6 {
		if _, err := tx.ExecContext(ctx, v6ToV7SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to widen the session_id column: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, createSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create eventstore_schema_version table: %w", err)
	}
	if _, err := tx.ExecContext(ctx, insertSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to insert initial schema version: %w", err)
	}
	if _, err := tx.ExecContext(ctx, dropSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to drop migrations table: %w", err)
	}

	return tx.Commit()
}
