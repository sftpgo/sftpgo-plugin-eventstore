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

// ProviderEvent defines a provider event
type ProviderEvent struct {
	ID         string `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Action     string `json:"action"`
	Username   string `json:"username"`
	IP         string `json:"ip,omitempty"`
	ObjectType string `json:"object_type"`
	ObjectName string `json:"object_name"`
	ObjectData []byte `json:"object_data"`
	Role       string `json:"role,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
}

func insertProviderEvent(ctx context.Context, ev *ProviderEvent) error {
	ev.ID = xid.New().String()
	q := buildInsertQuery("eventstore_provider_events", providerEventColumns)
	_, err := dbHandle.ExecContext(ctx, q,
		ev.ID, ev.Timestamp, ev.Action, ev.Username,
		ev.IP, ev.ObjectType, ev.ObjectName, ev.ObjectData,
		ev.Role, ev.InstanceID,
	)
	return err
}

func cleanupProviderEvents(timestamp time.Time) error {
	logger.AppLogger.Debug("removing provider events", "timestamp", timestamp)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	result, err := dbHandle.ExecContext(ctx, deleteQuery("eventstore_provider_events"),
		timestamp.UnixNano())
	if err == nil {
		deleted, _ := result.RowsAffected()
		logger.AppLogger.Debug("provider events deleted", "num", deleted)
	}
	return err
}
