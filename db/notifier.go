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
	"github.com/sftpgo/sdk/plugin/notifier"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

type Notifier struct {
	InstanceID string
}

func (n *Notifier) NotifyFsEvent(event *notifier.FsEvent) error {
	ev := &FsEvent{
		Timestamp:         event.Timestamp,
		Action:            event.Action,
		Username:          event.Username,
		FsPath:            event.Path,
		FsTargetPath:      event.TargetPath,
		VirtualPath:       event.VirtualPath,
		VirtualTargetPath: event.VirtualTargetPath,
		SSHCmd:            event.SSHCmd,
		Protocol:          event.Protocol,
		IP:                event.IP,
		SessionID:         event.SessionID,
		FileSize:          event.FileSize,
		Elapsed:           event.Elapsed,
		Status:            event.Status,
		FsProvider:        event.FsProvider,
		Bucket:            event.Bucket,
		Endpoint:          event.Endpoint,
		OpenFlags:         event.OpenFlags,
		Role:              event.Role,
		InstanceID:        n.InstanceID,
	}
	sess, cancel := GetDefaultSession()
	defer cancel()

	err := ev.Create(sess)
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
		Action:     event.Action,
		Username:   event.Username,
		IP:         event.IP,
		ObjectType: event.ObjectType,
		ObjectName: event.ObjectName,
		ObjectData: event.ObjectData,
		Role:       event.Role,
		InstanceID: n.InstanceID,
	}
	sess, cancel := GetDefaultSession()
	defer cancel()

	err := ev.Create(sess)
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
		Protocol:   event.Protocol,
		Username:   event.Username,
		IP:         event.IP,
		Message:    event.Message,
		Role:       event.Role,
		InstanceID: n.InstanceID,
	}
	sess, cancel := GetDefaultSession()
	defer cancel()

	err := ev.Create(sess)
	if err != nil {
		logger.AppLogger.Warn("unable to save log event", "event", event.Event, "error", err)
		return err
	}
	return nil
}
