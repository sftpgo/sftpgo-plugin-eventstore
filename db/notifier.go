package db

import (
	"time"

	"github.com/sftpgo/sftpgo-plugin-eventstore/logger"
)

type Notifier struct {
	InstanceID string
}

func (n *Notifier) NotifyFsEvent(timestamp time.Time, action, username, fsPath, fsTargetPath, sshCmd, protocol, ip,
	virtualPath, virtualTargetPath string, fileSize int64, status int,
) error {
	ev := &FsEvent{
		Timestamp:         timestamp,
		Action:            action,
		Username:          username,
		FsPath:            fsPath,
		FsTargetPath:      fsTargetPath,
		VirtualPath:       virtualPath,
		VirtualTargetPath: virtualTargetPath,
		SSHCmd:            sshCmd,
		Protocol:          protocol,
		IP:                ip,
		FileSize:          fileSize,
		Status:            status,
		InstanceID:        n.InstanceID,
	}
	sess, cancel := GetDefaultSession()
	defer cancel()

	err := ev.Create(sess)
	if err != nil {
		logger.AppLogger.Warn("unable to save fs event", "action", action, "username", username,
			"virtual path", virtualPath, "error", err)
		return err
	}
	return nil
}

func (n *Notifier) NotifyProviderEvent(timestamp time.Time, action, username, objectType, objectName, ip string,
	object []byte,
) error {
	ev := &ProviderEvent{
		Timestamp:  timestamp,
		Action:     action,
		Username:   username,
		IP:         ip,
		ObjectType: objectType,
		ObjectName: objectName,
		ObjectData: object,
		InstanceID: n.InstanceID,
	}
	sess, cancel := GetDefaultSession()
	defer cancel()

	err := ev.Create(sess)
	if err != nil {
		logger.AppLogger.Warn("unable to save provider event", "action", action, "error", err)
		return err
	}
	return nil
}
