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
		Status:            event.Status,
		FsProvider:        event.FsProvider,
		Bucket:            event.Bucket,
		Endpoint:          event.Endpoint,
		OpenFlags:         event.OpenFlags,
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
