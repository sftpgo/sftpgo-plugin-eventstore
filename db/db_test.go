package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/sftpgo/sftpgo-plugin-eventstore/db/migration"
)

func TestMain(m *testing.M) {
	driver := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_DRIVER")
	dsn := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_DSN")
	if driver == "" || dsn == "" {
		fmt.Println("Driver and/or DSN not set, unable to execute test")
		os.Exit(1)
	}
	if err := Initialize(driver, dsn, true); err != nil {
		fmt.Printf("unable to initialize database: %v\n", err)
		os.Exit(1)
	}
	if err := migration.MigrateDatabase(Handle); err != nil {
		fmt.Printf("unable to migrate database: %v\n", err)
		os.Exit(1)
	}
	exitCode := m.Run()
	os.Exit(exitCode)
}
