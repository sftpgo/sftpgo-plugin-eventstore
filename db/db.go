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
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

const (
	driverNamePostgreSQL = "postgres"
	driverNameMySQL      = "mysql"
)

var (
	dbHandle            *sql.DB
	driverName          string
	defaultQueryTimeout = 20 * time.Second
)

// Initialize initializes the database engine
func Initialize(driver, dsn, customTLSConfig string, poolSize int) error {
	var err error

	driverName = driver

	switch driverName {
	case driverNamePostgreSQL:
		dbHandle, err = sql.Open("pgx", dsn)
		if err != nil {
			logger.AppLogger.Error("unable to create db handle", "error", err)
			return err
		}
	case driverNameMySQL:
		if err := handleCustomTLSConfig(customTLSConfig); err != nil {
			logger.AppLogger.Error("unable to register custom tls config", "error", err)
			return err
		}
		dbHandle, err = sql.Open("mysql", dsn)
		if err != nil {
			logger.AppLogger.Error("unable to create db handle", "error", err)
			return err
		}
	default:
		return fmt.Errorf("unsupported database driver %q", driverName)
	}

	dbHandle.SetMaxOpenConns(poolSize)
	if poolSize > 0 {
		dbHandle.SetMaxIdleConns(poolSize)
	} else {
		dbHandle.SetMaxIdleConns(2)
	}
	dbHandle.SetConnMaxLifetime(240 * time.Second)
	dbHandle.SetConnMaxIdleTime(120 * time.Second)

	return dbHandle.Ping()
}

// Cleanup removes old events
func Cleanup(timestamp time.Time) {
	if err := cleanupFsEvents(timestamp); err != nil {
		logger.AppLogger.Error("unable to delete fs events", "error", err)
	}

	if err := cleanupProviderEvents(timestamp); err != nil {
		logger.AppLogger.Error("unable to delete provider events", "error", err)
	}

	if err := cleanupLogEvents(timestamp); err != nil {
		logger.AppLogger.Error("unable to delete log events", "error", err)
	}
}

func handleCustomTLSConfig(config string) error {
	if config == "" {
		return nil
	}
	values, err := url.ParseQuery(config)
	if err != nil {
		logger.AppLogger.Error("unable to parse custom tls config", "value", config, "error", err)
		return fmt.Errorf("unable to parse tls config: %w", err)
	}
	rootCert := values.Get("root_cert")
	clientCert := values.Get("client_cert")
	clientKey := values.Get("client_key")
	tlsMode := values.Get("tls_mode")

	tlsConfig := &tls.Config{}
	if rootCert != "" {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			rootCAs = x509.NewCertPool()
		}
		rootCrt, err := os.ReadFile(rootCert)
		if err != nil {
			return fmt.Errorf("unable to load root certificate %q: %v", rootCert, err)
		}
		if !rootCAs.AppendCertsFromPEM(rootCrt) {
			return fmt.Errorf("unable to parse root certificate %q", rootCert)
		}
		tlsConfig.RootCAs = rootCAs
	}
	if clientCert != "" && clientKey != "" {
		cert := make([]tls.Certificate, 0, 1)
		tlsCert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return fmt.Errorf("unable to load key pair %q, %q: %v", clientCert, clientKey, err)
		}
		cert = append(cert, tlsCert)
		tlsConfig.Certificates = cert
	}
	if tlsMode == "1" {
		tlsConfig.InsecureSkipVerify = true
	}

	if err := mysqldriver.RegisterTLSConfig("custom", tlsConfig); err != nil {
		return fmt.Errorf("unable to register tls config: %v", err)
	}
	return nil
}
