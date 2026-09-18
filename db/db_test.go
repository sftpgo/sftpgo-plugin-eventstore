// Copyright (C) 2021-2026 Nicola Murino
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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	driver := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_DRIVER")
	dsn := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_DSN")
	customTLSConfig := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_CUSTOM_TLS")
	if driver == "" || dsn == "" {
		fmt.Println("Driver and/or DSN not set, unable to execute test")
		os.Exit(1)
	}
	if err := Initialize(driver, dsn, customTLSConfig, 0); err != nil {
		fmt.Printf("unable to initialize database: %v\n", err)
		os.Exit(1)
	}
	if err := ResetDatabase(); err != nil {
		fmt.Printf("unable to reset database: %v\n", err)
		os.Exit(1)
	}
	if err := MigrateDatabase(); err != nil {
		fmt.Printf("unable to migrate database: %v\n", err)
		os.Exit(1)
	}
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestMigrationIdempotent(t *testing.T) {
	// calling MigrateDatabase again should be a no-op
	err := MigrateDatabase()
	assert.NoError(t, err)
}

func createGormigrateMarker(t *testing.T, id string) {
	t.Helper()
	ctx := context.Background()
	var createMigrations, insertMigration string
	switch driverName {
	case driverNameMySQL:
		createMigrations = "CREATE TABLE `migrations` (`id` varchar(191) PRIMARY KEY)"
		insertMigration = "INSERT INTO `migrations` (`id`) VALUES (?)"
	default:
		createMigrations = `CREATE TABLE "migrations" ("id" varchar(191) PRIMARY KEY)`
		insertMigration = `INSERT INTO "migrations" ("id") VALUES ($1)`
	}
	_, err := dbHandle.ExecContext(ctx, createMigrations)
	require.NoError(t, err)
	_, err = dbHandle.ExecContext(ctx, insertMigration, id)
	require.NoError(t, err)
}

func TestMigrationFromGormigrate(t *testing.T) {
	// gormigrate v7 install: schema v1 with no eventstore_schema_version,
	// "migrations" table marking version 7.
	ctx := context.Background()
	require.NoError(t, ResetDatabase())
	conn, err := dbHandle.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, initializeDatabase(ctx, conn))
	require.NoError(t, dropTable(ctx, conn, "eventstore_schema_version"))
	require.NoError(t, conn.Close())

	createGormigrateMarker(t, "7")
	require.NoError(t, MigrateDatabase())

	var version int
	err = dbHandle.QueryRowContext(ctx, "SELECT version FROM eventstore_schema_version LIMIT 1").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, schemaVersion, version)

	conn, err = dbHandle.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	assert.False(t, tableExists(ctx, conn, "migrations"))
}

func TestMigrationFromGormigratePartialRecovery(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, ResetDatabase())
	conn, err := dbHandle.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, initializeDatabase(ctx, conn))
	require.NoError(t, conn.Close())

	createGormigrateMarker(t, "7")

	require.NoError(t, MigrateDatabase())

	var version int
	err = dbHandle.QueryRowContext(ctx, "SELECT version FROM eventstore_schema_version LIMIT 1").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, schemaVersion, version)

	// Exactly one row in eventstore_schema_version (no duplicate insert).
	var rowCount int
	err = dbHandle.QueryRowContext(ctx, "SELECT COUNT(*) FROM eventstore_schema_version").Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)

	conn, err = dbHandle.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	assert.False(t, tableExists(ctx, conn, "migrations"))
}

func TestMigrationUnsupportedGormigrate(t *testing.T) {
	// gormigrate install at an old version (no "7" entry).
	ctx := context.Background()
	require.NoError(t, ResetDatabase())
	conn, err := dbHandle.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, initializeDatabase(ctx, conn))
	require.NoError(t, dropTable(ctx, conn, "eventstore_schema_version"))
	require.NoError(t, conn.Close())

	createGormigrateMarker(t, "5")

	err = MigrateDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported gormigrate version")

	conn, err = dbHandle.Conn(ctx)
	require.NoError(t, err)
	_ = dropTable(ctx, conn, "migrations")
	require.NoError(t, createSchemaVersionTable(ctx, conn, 1))
	require.NoError(t, conn.Close())
	require.NoError(t, MigrateDatabase())
}

func TestResetAndFreshInstall(t *testing.T) {
	// reset should drop all tables
	err := ResetDatabase()
	require.NoError(t, err)

	// migrate should do a fresh install
	err = MigrateDatabase()
	require.NoError(t, err)

	// verify eventstore_schema_version
	var version int
	ctx := context.Background()
	err = dbHandle.QueryRowContext(ctx, "SELECT version FROM eventstore_schema_version LIMIT 1").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, schemaVersion, version)
}

func TestUnsupportedDriver(t *testing.T) {
	currentDriver := driverName
	currentHandle := dbHandle
	t.Cleanup(func() {
		driverName = currentDriver
		dbHandle = currentHandle
	})

	err := Initialize("unsupported", "dsn", "", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

func TestBuildInsertQuery(t *testing.T) {
	currentDriver := driverName
	t.Cleanup(func() {
		driverName = currentDriver
	})

	driverName = driverNamePostgreSQL
	assert.Equal(t, `INSERT INTO tbl ("a","b") VALUES ($1,$2)`,
		buildInsertQuery("tbl", []string{"a", "b"}))
	assert.Equal(t, `DELETE FROM tbl WHERE "timestamp" < $1`, deleteQuery("tbl"))

	driverName = driverNameMySQL
	assert.Equal(t, "INSERT INTO tbl (`a`,`b`) VALUES (?,?)",
		buildInsertQuery("tbl", []string{"a", "b"}))
	assert.Equal(t, "DELETE FROM tbl WHERE `timestamp` < ?", deleteQuery("tbl"))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFsEventRow(row rowScanner) (FsEvent, error) {
	var ev FsEvent
	var fsPath, fsTargetPath, virtualPath, virtualTargetPath sql.NullString
	var sshCmd, ip, sessionID, bucket, endpoint, role, instanceID sql.NullString
	var fileSize, elapsed sql.NullInt64
	var status, fsProvider, openFlags sql.NullInt32

	err := row.Scan(
		&ev.ID, &ev.Timestamp, &ev.Action, &ev.Username,
		&fsPath, &fsTargetPath, &virtualPath, &virtualTargetPath,
		&sshCmd, &fileSize, &elapsed, &status,
		&ev.Protocol, &ip, &sessionID, &fsProvider,
		&bucket, &endpoint, &openFlags, &role, &instanceID,
	)
	if err != nil {
		return ev, err
	}
	ev.FsPath = fsPath.String
	ev.FsTargetPath = fsTargetPath.String
	ev.VirtualPath = virtualPath.String
	ev.VirtualTargetPath = virtualTargetPath.String
	ev.SSHCmd = sshCmd.String
	ev.FileSize = fileSize.Int64
	ev.Elapsed = elapsed.Int64
	ev.Status = int(status.Int32)
	ev.IP = ip.String
	ev.SessionID = sessionID.String
	ev.FsProvider = int(fsProvider.Int32)
	ev.Bucket = bucket.String
	ev.Endpoint = endpoint.String
	ev.OpenFlags = int(openFlags.Int32)
	ev.Role = role.String
	ev.InstanceID = instanceID.String
	return ev, nil
}

func scanProviderEventRow(row rowScanner) (ProviderEvent, error) {
	var ev ProviderEvent
	var ip, objectType, objectName, role, instanceID sql.NullString

	err := row.Scan(
		&ev.ID, &ev.Timestamp, &ev.Action, &ev.Username,
		&ip, &objectType, &objectName, &ev.ObjectData,
		&role, &instanceID,
	)
	if err != nil {
		return ev, err
	}
	ev.IP = ip.String
	ev.ObjectType = objectType.String
	ev.ObjectName = objectName.String
	ev.Role = role.String
	ev.InstanceID = instanceID.String
	return ev, nil
}

func scanLogEventRow(row rowScanner) (LogEvent, error) {
	var ev LogEvent
	var protocol, username, ip, message, role, instanceID sql.NullString

	err := row.Scan(
		&ev.ID, &ev.Timestamp, &ev.Event, &protocol,
		&username, &ip, &message, &role, &instanceID,
	)
	if err != nil {
		return ev, err
	}
	ev.Protocol = protocol.String
	ev.Username = username.String
	ev.IP = ip.String
	ev.Message = message.String
	ev.Role = role.String
	ev.InstanceID = instanceID.String
	return ev, nil
}

func TestRecoverInterruptedInstall(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, ResetDatabase())

	// initializeDatabase leaves exactly the initial schema; dropping the
	// version row reproduces the interrupted state.
	conn, err := dbHandle.Conn(ctx)
	require.NoError(t, err)
	require.NoError(t, initializeDatabase(ctx, conn))
	require.NoError(t, conn.Close())
	_, err = dbHandle.ExecContext(ctx, "DELETE FROM eventstore_schema_version")
	require.NoError(t, err)

	// the wedge: this used to fail here, and on every subsequent run
	require.NoError(t, MigrateDatabase())

	var version, rows int
	require.NoError(t, dbHandle.QueryRowContext(ctx,
		"SELECT version FROM eventstore_schema_version LIMIT 1").Scan(&version))
	assert.Equal(t, schemaVersion, version)
	require.NoError(t, dbHandle.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM eventstore_schema_version").Scan(&rows))
	assert.Equal(t, 1, rows, "recovery must not duplicate the version row")

	// and it stays stable when run again
	require.NoError(t, MigrateDatabase())
	require.NoError(t, dbHandle.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM eventstore_schema_version").Scan(&rows))
	assert.Equal(t, 1, rows)
}
