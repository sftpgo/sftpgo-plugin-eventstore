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
	customTLSConfig := os.Getenv("SFTPGO_PLUGIN_EVENTSTORE_CUSTOM_TLS")
	if driver == "" || dsn == "" {
		fmt.Println("Driver and/or DSN not set, unable to execute test")
		os.Exit(1)
	}
	if err := Initialize(driver, dsn, customTLSConfig, true); err != nil {
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
