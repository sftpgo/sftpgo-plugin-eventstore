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
	"fmt"
	"strings"
)

var (
	fsEventColumns = []string{
		"id", "timestamp", "action", "username",
		"fs_path", "fs_target_path", "virtual_path", "virtual_target_path",
		"ssh_cmd", "file_size", "elapsed", "status",
		"protocol", "ip", "session_id", "fs_provider",
		"bucket", "endpoint", "open_flags", "role", "instance_id",
	}
	providerEventColumns = []string{
		"id", "timestamp", "action", "username",
		"ip", "object_type", "object_name", "object_data",
		"role", "instance_id",
	}
	logEventColumns = []string{
		"id", "timestamp", "event", "protocol",
		"username", "ip", "message", "role", "instance_id",
	}
)

func deleteQuery(table string) string {
	if driverName == driverNamePostgreSQL {
		return fmt.Sprintf("DELETE FROM %s WHERE %s < $1", table, quoteColumn("timestamp"))
	}
	return fmt.Sprintf("DELETE FROM %s WHERE %s < ?", table, quoteColumn("timestamp"))
}

func quoteColumn(col string) string {
	switch driverName {
	case driverNameMySQL:
		return "`" + col + "`"
	default:
		return `"` + col + `"`
	}
}

func quoteColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = quoteColumn(c)
	}
	return strings.Join(quoted, ",")
}

func buildInsertQuery(table string, columns []string) string {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	sb.WriteString(quoteColumns(columns))
	sb.WriteString(") VALUES (")
	if driverName == driverNamePostgreSQL {
		for i := range columns {
			if i > 0 {
				sb.WriteString(",")
			}
			fmt.Fprintf(&sb, "$%d", i+1)
		}
	} else {
		sb.WriteString(strings.TrimRight(strings.Repeat("?,", len(columns)), ","))
	}
	sb.WriteString(")")
	return sb.String()
}

const (
	mysqlInitialSQL = `CREATE TABLE IF NOT EXISTS eventstore_fs_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  timestamp bigint NOT NULL,
  action varchar(60) NOT NULL,
  username varchar(255) NOT NULL,
  fs_path longtext,
  fs_target_path longtext,
  virtual_path longtext,
  virtual_target_path longtext,
  ssh_cmd varchar(60),
  file_size bigint,
  elapsed bigint,
  status int,
  protocol varchar(30) NOT NULL,
  ip varchar(50),
  session_id varchar(512),
  fs_provider int,
  bucket varchar(512),
  endpoint varchar(512),
  open_flags int,
  role varchar(255),
  instance_id varchar(60),
  KEY idx_fs_events_timestamp (timestamp),
  KEY idx_fs_events_action (action),
  KEY idx_fs_events_username (username),
  KEY idx_fs_events_ssh_cmd (ssh_cmd),
  KEY idx_fs_events_status (status),
  KEY idx_fs_events_protocol (protocol),
  KEY idx_fs_events_ip (ip),
  KEY idx_fs_events_provider (fs_provider),
  KEY idx_fs_events_bucket (bucket),
  KEY idx_fs_events_endpoint (endpoint),
  KEY idx_fs_events_role (role),
  KEY idx_fs_events_instance_id (instance_id)
);
CREATE TABLE IF NOT EXISTS eventstore_provider_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  timestamp bigint NOT NULL,
  action varchar(60) NOT NULL,
  username varchar(255) NOT NULL,
  ip varchar(50),
  object_type varchar(50),
  object_name varchar(255),
  object_data longblob,
  role varchar(255),
  instance_id varchar(60),
  KEY idx_provider_events__timestamp (timestamp),
  KEY idx_provider_events_action (action),
  KEY idx_provider_events_username (username),
  KEY idx_provider_events_ip (ip),
  KEY idx_provider_events_object_type (object_type),
  KEY idx_provider_events_object_name (object_name),
  KEY idx_provider_events_role (role),
  KEY idx_provider_events_instance_id (instance_id)
);
CREATE TABLE IF NOT EXISTS eventstore_log_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  timestamp bigint NOT NULL,
  event int NOT NULL,
  protocol varchar(30),
  username varchar(255),
  ip varchar(50),
  message longtext,
  role varchar(255),
  instance_id varchar(60),
  KEY idx_log_events_timestamp (timestamp),
  KEY idx_log_events_event (event),
  KEY idx_log_events_protocol (protocol),
  KEY idx_log_events_username (username),
  KEY idx_log_events_ip (ip),
  KEY idx_log_events_role (role),
  KEY idx_log_events_instance_id (instance_id)
);
CREATE TABLE IF NOT EXISTS eventstore_schema_version (
  id int AUTO_INCREMENT NOT NULL PRIMARY KEY,
  version int NOT NULL
);
INSERT INTO eventstore_schema_version (version) SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM eventstore_schema_version)`

	pgsqlInitialSQL = `CREATE TABLE IF NOT EXISTS eventstore_fs_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  "timestamp" bigint NOT NULL,
  action varchar(60) NOT NULL,
  username varchar(255) NOT NULL,
  fs_path text,
  fs_target_path text,
  virtual_path text,
  virtual_target_path text,
  ssh_cmd varchar(60),
  file_size bigint,
  elapsed bigint,
  status int,
  protocol varchar(30) NOT NULL,
  ip varchar(50),
  session_id varchar(512),
  fs_provider int,
  bucket varchar(512),
  endpoint varchar(512),
  open_flags int,
  role varchar(255),
  instance_id varchar(60)
);
CREATE INDEX IF NOT EXISTS idx_fs_events_timestamp ON eventstore_fs_events ("timestamp");
CREATE INDEX IF NOT EXISTS idx_fs_events_action ON eventstore_fs_events (action);
CREATE INDEX IF NOT EXISTS idx_fs_events_username ON eventstore_fs_events (username);
CREATE INDEX IF NOT EXISTS idx_fs_events_ssh_cmd ON eventstore_fs_events (ssh_cmd);
CREATE INDEX IF NOT EXISTS idx_fs_events_status ON eventstore_fs_events (status);
CREATE INDEX IF NOT EXISTS idx_fs_events_protocol ON eventstore_fs_events (protocol);
CREATE INDEX IF NOT EXISTS idx_fs_events_ip ON eventstore_fs_events (ip);
CREATE INDEX IF NOT EXISTS idx_fs_events_provider ON eventstore_fs_events (fs_provider);
CREATE INDEX IF NOT EXISTS idx_fs_events_bucket ON eventstore_fs_events (bucket);
CREATE INDEX IF NOT EXISTS idx_fs_events_endpoint ON eventstore_fs_events (endpoint);
CREATE INDEX IF NOT EXISTS idx_fs_events_role ON eventstore_fs_events (role);
CREATE INDEX IF NOT EXISTS idx_fs_events_instance_id ON eventstore_fs_events (instance_id);
CREATE TABLE IF NOT EXISTS eventstore_provider_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  "timestamp" bigint NOT NULL,
  action varchar(60) NOT NULL,
  username varchar(255) NOT NULL,
  ip varchar(50),
  object_type varchar(50),
  object_name varchar(255),
  object_data bytea,
  role varchar(255),
  instance_id varchar(60)
);
CREATE INDEX IF NOT EXISTS idx_provider_events__timestamp ON eventstore_provider_events ("timestamp");
CREATE INDEX IF NOT EXISTS idx_provider_events_action ON eventstore_provider_events (action);
CREATE INDEX IF NOT EXISTS idx_provider_events_username ON eventstore_provider_events (username);
CREATE INDEX IF NOT EXISTS idx_provider_events_ip ON eventstore_provider_events (ip);
CREATE INDEX IF NOT EXISTS idx_provider_events_object_type ON eventstore_provider_events (object_type);
CREATE INDEX IF NOT EXISTS idx_provider_events_object_name ON eventstore_provider_events (object_name);
CREATE INDEX IF NOT EXISTS idx_provider_events_role ON eventstore_provider_events (role);
CREATE INDEX IF NOT EXISTS idx_provider_events_instance_id ON eventstore_provider_events (instance_id);
CREATE TABLE IF NOT EXISTS eventstore_log_events (
  id varchar(36) NOT NULL PRIMARY KEY,
  "timestamp" bigint NOT NULL,
  event int NOT NULL,
  protocol varchar(30),
  username varchar(255),
  ip varchar(50),
  message text,
  role varchar(255),
  instance_id varchar(60)
);
CREATE INDEX IF NOT EXISTS idx_log_events_timestamp ON eventstore_log_events ("timestamp");
CREATE INDEX IF NOT EXISTS idx_log_events_event ON eventstore_log_events (event);
CREATE INDEX IF NOT EXISTS idx_log_events_protocol ON eventstore_log_events (protocol);
CREATE INDEX IF NOT EXISTS idx_log_events_username ON eventstore_log_events (username);
CREATE INDEX IF NOT EXISTS idx_log_events_ip ON eventstore_log_events (ip);
CREATE INDEX IF NOT EXISTS idx_log_events_role ON eventstore_log_events (role);
CREATE INDEX IF NOT EXISTS idx_log_events_instance_id ON eventstore_log_events (instance_id);
CREATE TABLE IF NOT EXISTS eventstore_schema_version (
  id integer NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  version int NOT NULL
);
INSERT INTO eventstore_schema_version (version) SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM eventstore_schema_version)`
)
