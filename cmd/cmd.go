package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/sftpgo/sdk/plugin/notifier"
	"github.com/urfave/cli/v2"

	"github.com/sftpgo/sftpgo-plugin-eventstore/db"
	"github.com/sftpgo/sftpgo-plugin-eventstore/db/migration"
	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

const (
	version   = "1.0.6"
	envPrefix = "SFTPGO_PLUGIN_EVENTSTORE_"
)

var (
	commitHash = ""
	buildDate  = ""
)

var (
	driver     string
	instanceID string
	dsn        string
	retention  int

	dbFlags = []cli.Flag{
		&cli.StringFlag{
			Name:        "driver",
			Usage:       "Database driver (required)",
			Destination: &driver,
			EnvVars:     []string{envPrefix + "DRIVER"},
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "dsn",
			Usage:       "Data source URI (required)",
			Destination: &dsn,
			EnvVars:     []string{envPrefix + "DSN"},
			Required:    true,
		},
	}

	serveFlags = append(dbFlags,
		&cli.StringFlag{
			Name:        "instance-id",
			Usage:       "Instance identifier",
			Destination: &instanceID,
			EnvVars:     []string{envPrefix + "INSTANCE_ID"},
		},
		&cli.IntFlag{
			Name:        "retention",
			Usage:       `Events older than the specified number of hours will be deleted. 0 means no events will be deleted`,
			Destination: &retention,
			EnvVars:     []string{envPrefix + "RETENTION"},
		},
	)

	rootCmd = &cli.App{
		Name:    "sftpgo-plugin-eventstore",
		Version: getVersionString(),
		Usage:   "SFTPGo events store plugin",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Launch the SFTPGo plugin, it must be called from an SFTPGo instance",
				Flags: serveFlags,
				Action: func(c *cli.Context) error {
					logger.AppLogger.Info("starting sftpgo-plugin-eventstore", "version", getVersionString(),
						"database driver", driver, "instance id", instanceID)
					if err := db.Initialize(driver, dsn, false); err != nil {
						logger.AppLogger.Error("unable to initialize database", "error", err)
						return err
					}
					if err := migration.MigrateDatabase(db.Handle); err != nil {
						logger.AppLogger.Error("unable to migrate database", "error", err)
						return err
					}
					if retention > 0 {
						go dbCleanup(retention)
					} else {
						logger.AppLogger.Debug("retention not set, no event will be deleted")
					}

					plugin.Serve(&plugin.ServeConfig{
						HandshakeConfig: notifier.Handshake,
						Plugins: map[string]plugin.Plugin{
							notifier.PluginName: &notifier.Plugin{Impl: &db.Notifier{
								InstanceID: instanceID,
							}},
						},
						GRPCServer: plugin.DefaultGRPCServer,
					})

					return errors.New("the plugin exited unexpectedly")
				},
			},
			{
				Name:  "migrate",
				Usage: "Apply database schema migrations",
				Flags: dbFlags,
				Action: func(c *cli.Context) error {
					if err := db.Initialize(driver, dsn, true); err != nil {
						logger.AppLogger.Error("unable to initialize database", "error", err)
						return err
					}
					if err := migration.MigrateDatabase(db.Handle); err != nil {
						logger.AppLogger.Error("unable to migrate database", "error", err)
						return err
					}
					return nil
				},
			},
			{
				Name:  "reset",
				Usage: "Reset the database schema, any data will be lost",
				Flags: dbFlags,
				Action: func(c *cli.Context) error {
					fmt.Println("You are about to delete all database data and schema", "driver", fmt.Sprintf("%#v", driver),
						"dsn", fmt.Sprintf("%#v", dsn), "Are you sure?")
					fmt.Println("Y/n")
					reader := bufio.NewReader(os.Stdin)
					answer, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("unexpected error", err)
						return err
					}
					if strings.ToUpper(strings.TrimSpace(answer)) != "Y" {
						fmt.Println("Aborted!")
						return errors.New("command aborted")
					}
					if err := db.Initialize(driver, dsn, true); err != nil {
						logger.AppLogger.Error("unable to initialize database", "error", err)
						return err
					}
					if err := migration.ResetDatabase(db.Handle); err != nil {
						logger.AppLogger.Error("unable to reset database", "error", err)
						return err
					}
					return nil
				},
			},
		},
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Run(os.Args)
}

func dbCleanup(retentionHours int) {
	logger.AppLogger.Debug("start event retention check, old events will be checked every hour",
		"retention (hours)", retentionHours)
	for range time.Tick(1 * time.Hour) {
		db.Cleanup(time.Now().Add(-time.Duration(retentionHours) * time.Hour))
	}
}

func getVersionString() string {
	var sb strings.Builder
	sb.WriteString(version)
	if commitHash != "" {
		sb.WriteString("-")
		sb.WriteString(commitHash)
	}
	if buildDate != "" {
		sb.WriteString("-")
		sb.WriteString(buildDate)
	}
	return sb.String()
}
