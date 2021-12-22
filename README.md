# SFTPGo events store plugin

![Build](https://github.com/sftpgo/sftpgo-plugin-eventstore/workflows/Build/badge.svg?branch=main&event=push)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

This plugin allows to store [SFTPGo](https://github.com/drakkan/sftpgo/) filesystem and provider events in supported database engines. It is not meant to react to `pre-*` events.

## Configuration

The plugin can be configured within the `plugins` section of the SFTPGo configuration file. To start the plugin you have to use the `serve` subcommand. Here is the usage.

```shell
NAME:
   sftpgo-plugin-eventstore serve - Launch the SFTPGo plugin, it must be called from an SFTPGo instance

USAGE:
   sftpgo-plugin-eventstore serve [command options] [arguments...]

OPTIONS:
   --driver value       Database driver (required) [$SFTPGO_PLUGIN_EVENTSTORE_DRIVER]
   --dsn value          Data source URI (required) [$SFTPGO_PLUGIN_EVENTSTORE_DSN]
   --instance-id value  Instance identifier [$SFTPGO_PLUGIN_EVENTSTORE_INSTANCE_ID]
   --retention value    Events older than the specified number of hours will be deleted. 0 means no events will be deleted (default: 0) [$SFTPGO_PLUGIN_EVENTSTORE_RETENTION]
```

The `driver` and `dsn` flags are required. The `instance-id` allows to set an identifier, it is useful if you are storing events from multiple SFTPGo instances and want to store where they are coming from.
If you set a `retention` > 0 events older than `now - retention (in hours)` will be automatically deleted. Old events will be checked every hour.
Each flag can also be set using environment variables, for example the DSN can be set using the `SFTPGO_PLUGIN_EVENTSTORE_DSN` environment variable.

This is an example configuration.

```json
...
"plugins": [
    {
      "type": "notifier",
      "notifier_options": {
        "fs_events": [
          "download",
          "upload",
          "delete",
          "rename",
          "mkdir",
          "rmdir",
          "ssh_cmd"
        ],
        "provider_events": [
          "add",
          "update",
          "delete"
        ],
        "provider_objects": [
          "user",
          "admin",
          "api_key"
        ],
        "retry_max_time": 60,
        "retry_queue_max_size": 1000
      },
      "cmd": "<path to sftpgo-plugin-eventstore>",
      "args": ["serve", "--driver", "postgres"],
      "sha256sum": "",
      "auto_mtls": true
    }
  ]
...
```

With the above example the plugin is configured to connect to PostgreSQL. We set the DSN using the `SFTPGO_PLUGIN_EVENTSTORE_DSN` environment variable.

The plugin will not start if it fails to connect to the configured database service, this will prevent SFTPGo from starting.

SFTPGo will automatically restart it if it crashes and you can configure SFTPGo to retry failed events until they are older than a configurable time (60 seconds in the above example). This way no event is lost.

The plugin supports also the `migrate` and `reset` sub-commands that can be used in standalone mode and are useful for debugging purposes. Please refer to their help texts for usage.

## Database tables

The plugin will automatically create the following database tables:

- `eventstore_fs_events`
- `eventstore_provider_events`

Inspect your database for more details.

## Supported database services

### PostgreSQL

To use Postgres you have to use `postgres` as driver. If you have a database named `sftpgo_events` on localhost and you want to connect to it using the user `sftpgo` with the password `sftpgopass` you can use a DSN like the following one.

```shell
"host='127.0.0.1' port=5432 dbname='sftpgo_events' user='sftpgo' password='sftpgopass' sslmode=disable connect_timeout=10"
```

Please refer to the documentation [here](https://github.com/go-gorm/postgres) for details about the dsn.

### MySQL

To use MySQL you have to use `mysql` as driver. If you have a database named `sftpgo_events` on localhost and you want to connect to it using the user `sftpgo` with the password `sftpgopass` you can use a DSN like the following one.

```shell
"sftpgo:sftpgopass@tcp([127.0.0.1]:3306)/sftpgo_events?charset=utf8mb4&interpolateParams=true&timeout=10s&tls=false&writeTimeout=10s&readTimeout=10s&parseTime=true"
```

Please refer to the documentation [here](https://github.com/go-gorm/mysql) for details about the dsn.
