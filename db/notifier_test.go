package db

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/sftpgo/sdk/plugin/notifier"
	"github.com/stretchr/testify/assert"
)

func TestNotifyEvents(t *testing.T) {
	n := Notifier{
		InstanceID: "sftpgo1",
	}

	fsEvent := &notifier.FsEvent{
		Timestamp:         time.Now().UnixNano(),
		Action:            "upload",
		Username:          "username",
		Path:              "/tmp/file.txt",
		TargetPath:        "/tmp/target.txt",
		VirtualPath:       "file.txt",
		VirtualTargetPath: "target.txt",
		SSHCmd:            "scp",
		FileSize:          123,
		Elapsed:           456,
		Status:            1,
		Protocol:          "SFTP",
		SessionID:         xid.New().String(),
		IP:                "::1",
		FsProvider:        1,
		Bucket:            "bucket",
		Endpoint:          "endpoint",
		Role:              "role1",
		OpenFlags:         512,
	}

	err := n.NotifyFsEvent(fsEvent)
	assert.NoError(t, err)

	sess, cancel := GetDefaultSession()
	defer cancel()

	var event FsEvent
	err = sess.First(&event).Error
	assert.NoError(t, err)

	assert.Equal(t, n.InstanceID, event.InstanceID)
	assert.NotEmpty(t, event.ID)

	assert.Equal(t, fsEvent.Timestamp, event.Timestamp)
	assert.Equal(t, fsEvent.Action, event.Action)
	assert.Equal(t, fsEvent.Username, event.Username)
	assert.Equal(t, fsEvent.Path, event.FsPath)
	assert.Equal(t, fsEvent.TargetPath, event.FsTargetPath)
	assert.Equal(t, fsEvent.VirtualPath, event.VirtualPath)
	assert.Equal(t, fsEvent.VirtualTargetPath, event.VirtualTargetPath)
	assert.Equal(t, fsEvent.SSHCmd, event.SSHCmd)
	assert.Equal(t, fsEvent.FileSize, event.FileSize)
	assert.Equal(t, fsEvent.Elapsed, event.Elapsed)
	assert.Equal(t, fsEvent.Status, event.Status)
	assert.Equal(t, fsEvent.Protocol, event.Protocol)
	assert.Equal(t, fsEvent.SessionID, event.SessionID)
	assert.Equal(t, fsEvent.IP, event.IP)
	assert.Equal(t, fsEvent.FsProvider, event.FsProvider)
	assert.Equal(t, fsEvent.Bucket, event.Bucket)
	assert.Equal(t, fsEvent.Endpoint, event.Endpoint)
	assert.Equal(t, fsEvent.OpenFlags, event.OpenFlags)
	assert.Equal(t, fsEvent.Role, event.Role)

	providerEvent := &notifier.ProviderEvent{
		Timestamp:  time.Now().UnixNano(),
		Action:     "add",
		Username:   "adminUsername",
		IP:         "127.0.0.1",
		ObjectType: "admin",
		ObjectName: "adminname",
		Role:       "role2",
		ObjectData: []byte("data"),
	}

	err = n.NotifyProviderEvent(providerEvent)
	assert.NoError(t, err)

	var providerEv ProviderEvent
	err = sess.First(&providerEv).Error
	assert.NoError(t, err)

	assert.Equal(t, n.InstanceID, providerEv.InstanceID)
	assert.NotEmpty(t, providerEv.ID)

	assert.Equal(t, providerEvent.Timestamp, providerEv.Timestamp)
	assert.Equal(t, providerEvent.Action, providerEv.Action)
	assert.Equal(t, providerEvent.Username, providerEv.Username)
	assert.Equal(t, providerEvent.IP, providerEv.IP)
	assert.Equal(t, providerEvent.ObjectType, providerEv.ObjectType)
	assert.Equal(t, providerEvent.ObjectName, providerEv.ObjectName)
	assert.Equal(t, providerEvent.Role, providerEv.Role)
	assert.Equal(t, providerEvent.ObjectData, providerEv.ObjectData)

	logEvent := &notifier.LogEvent{
		Timestamp: time.Now().UnixNano(),
		Event:     1,
		Protocol:  "SSH",
		Username:  "user1",
		IP:        "127.0.0.1",
		Message:   "error desc",
		Role:      "role1",
	}
	err = n.NotifyLogEvent(logEvent)
	assert.NoError(t, err)

	var logEv LogEvent
	err = sess.First(&logEv).Error
	assert.NoError(t, err)

	assert.Equal(t, n.InstanceID, logEv.InstanceID)
	assert.NotEmpty(t, logEv.ID)
	assert.Equal(t, logEvent.Timestamp, logEv.Timestamp)
	assert.Equal(t, int(logEvent.Event), logEv.Event)
	assert.Equal(t, logEvent.Protocol, logEv.Protocol)
	assert.Equal(t, logEvent.Username, logEv.Username)
	assert.Equal(t, logEvent.IP, logEv.IP)
	assert.Equal(t, logEvent.Message, logEv.Message)
	assert.Equal(t, logEvent.Role, logEv.Role)

	// test cleanup
	Cleanup(time.Now().Add(-24 * time.Hour))
	// the data must not be deleted
	var fsEvents []FsEvent
	result := sess.Find(&fsEvents)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)

	var providerEvents []ProviderEvent
	result = sess.Find(&providerEvents)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)

	fsEvents = nil
	providerEvents = nil
	Cleanup(time.Now().Add(1 * time.Hour))
	result = sess.Find(&fsEvents)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(0), result.RowsAffected)

	result = sess.Find(&providerEvents)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(0), result.RowsAffected)
}
