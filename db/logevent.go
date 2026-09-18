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
	"time"

	"github.com/rs/xid"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

// LogEvent defines a log event
type LogEvent struct {
	ID         string `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Event      int    `json:"event"`
	Protocol   string `json:"protocol,omitempty"`
	Username   string `json:"username,omitempty"`
	IP         string `json:"ip,omitempty"`
	Message    string `json:"message,omitempty"`
	Role       string `json:"role,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
}

func insertLogEvent(ctx context.Context, ev *LogEvent) error {
	ev.ID = xid.New().String()
	q := buildInsertQuery("eventstore_log_events", logEventColumns)
	_, err := dbHandle.ExecContext(ctx, q,
		ev.ID, ev.Timestamp, ev.Event, ev.Protocol,
		ev.Username, ev.IP, ev.Message, ev.Role, ev.InstanceID,
	)
	return err
}

func cleanupLogEvents(timestamp time.Time) error {
	logger.AppLogger.Debug("removing log events", "timestamp", timestamp)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	result, err := dbHandle.ExecContext(ctx, deleteQuery("eventstore_log_events"),
		timestamp.UnixNano())
	if err == nil {
		deleted, _ := result.RowsAffected()
		logger.AppLogger.Debug("log events deleted", "num", deleted)
	}
	return err
}
