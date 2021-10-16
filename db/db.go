package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

const (
	driverNamePostgreSQL = "postgres"
	driverNameMySQL      = "mysql"
)

var (
	// Handle defines the global database handle
	Handle              *gorm.DB
	defaultQueryTimeout = 20 * time.Second
	driverName          string
)

// Initialize initializes the database engine
func Initialize(driver, dsn string, dbDebug bool) error {
	var err error

	newLogger := gormlogger.Discard

	if dbDebug {
		newLogger = gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			gormlogger.Config{
				SlowThreshold: time.Second,     // Slow SQL threshold
				LogLevel:      gormlogger.Info, // Log level
				Colorful:      runtime.GOOS != "windows",
			},
		)
	}

	driverName = driver

	switch driverName {
	case driverNamePostgreSQL:
		Handle, err = gorm.Open(postgres.New(postgres.Config{
			DSN: dsn,
		}), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 newLogger,
		})
		if err != nil {
			logger.AppLogger.Error("unable to create db handle", "error", err)
			return err
		}
	case driverNameMySQL:
		Handle, err = gorm.Open(mysql.New(mysql.Config{
			DSN: dsn,
		}), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 newLogger,
		})
		if err != nil {
			logger.AppLogger.Error("unable to create db handle", "error", err)
			return err
		}
	default:
		return fmt.Errorf("unsupported database driver %v", driverName)
	}

	sqlDB, err := Handle.DB()
	if err != nil {
		logger.AppLogger.Error("unable to get sql db handle", "error", err)
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)

	return sqlDB.Ping()
}

// GetDefaultSession returns a database session with the default timeout.
// Don't forget to cancel the returned context
func GetDefaultSession() (*gorm.DB, context.CancelFunc) {
	return getSessionWithTimeout(defaultQueryTimeout)
}

// getSessionWithTimeout returns a database session with the specified timeout.
// Don't forget to cancel the returned context
func getSessionWithTimeout(timeout time.Duration) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return Handle.WithContext(ctx), cancel
}

// Cleanup removes old events
func Cleanup(timestamp time.Time) {
	if err := cleanupFsEvents(timestamp); err != nil {
		logger.AppLogger.Error("unable to delete fs events", "error", err)
	}

	if err := cleanupProviderEvents(timestamp); err != nil {
		logger.AppLogger.Error("unable to delete provider events", "error", err)
	}
}
