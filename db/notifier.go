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
	"strings"

	"github.com/sftpgo/sdk/plugin/notifier"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

type Notifier struct {
	InstanceID string
}

func (n *Notifier) NotifyFsEvent(event *notifier.FsEvent) error {
	ev := &FsEvent{
		Timestamp:         event.Timestamp,
		Action:            sanitizeText(event.Action),
		Username:          sanitizeText(event.Username),
		FsPath:            sanitizeText(event.Path),
		FsTargetPath:      sanitizeText(event.TargetPath),
		VirtualPath:       sanitizeText(event.VirtualPath),
		VirtualTargetPath: sanitizeText(event.VirtualTargetPath),
		SSHCmd:            sanitizeText(event.SSHCmd),
		Protocol:          sanitizeText(event.Protocol),
		IP:                sanitizeText(event.IP),
		SessionID:         sanitizeText(event.SessionID),
		FileSize:          event.FileSize,
		Elapsed:           event.Elapsed,
		Status:            event.Status,
		FsProvider:        event.FsProvider,
		Bucket:            sanitizeText(event.Bucket),
		Endpoint:          sanitizeText(event.Endpoint),
		OpenFlags:         event.OpenFlags,
		Role:              sanitizeText(event.Role),
		InstanceID:        n.InstanceID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	err := insertFsEvent(ctx, ev)
	if err != nil {
		logger.AppLogger.Warn("unable to save fs event", "action", event.Action, "username",
			event.Username, "virtual path", event.VirtualPath, "error", err)
		return err
	}
	return nil
}

func (n *Notifier) NotifyProviderEvent(event *notifier.ProviderEvent) error {
	ev := &ProviderEvent{
		Timestamp:  event.Timestamp,
		Action:     sanitizeText(event.Action),
		Username:   sanitizeText(event.Username),
		IP:         sanitizeText(event.IP),
		ObjectType: sanitizeText(event.ObjectType),
		ObjectName: sanitizeText(event.ObjectName),
		ObjectData: event.ObjectData,
		Role:       sanitizeText(event.Role),
		InstanceID: n.InstanceID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	err := insertProviderEvent(ctx, ev)
	if err != nil {
		logger.AppLogger.Warn("unable to save provider event", "action", event.Action, "error", err)
		return err
	}
	return nil
}

func (n *Notifier) NotifyLogEvent(event *notifier.LogEvent) error {
	ev := &LogEvent{
		Timestamp:  event.Timestamp,
		Event:      int(event.Event),
		Protocol:   sanitizeText(event.Protocol),
		Username:   sanitizeText(event.Username),
		IP:         sanitizeText(event.IP),
		Message:    sanitizeText(event.Message),
		Role:       sanitizeText(event.Role),
		InstanceID: n.InstanceID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()

	err := insertLogEvent(ctx, ev)
	if err != nil {
		logger.AppLogger.Warn("unable to save log event", "event", event.Event, "error", err)
		return err
	}
	return nil
}

func sanitizeText(s string) string {
	return strings.ReplaceAll(strings.ToValidUTF8(s, "\uFFFD"), "\x00", "\uFFFD")
}
