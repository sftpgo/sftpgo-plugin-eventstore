package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sftpgo/sdk/plugin/notifier"
	"github.com/stretchr/testify/assert"
)

func TestFsEvent(t *testing.T) {
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
		Status:            1,
		Protocol:          "SFTP",
		SessionID:         uuid.NewString(),
		IP:                "::1",
		FsProvider:        1,
		Bucket:            "bucket",
		Endpoint:          "endpoint",
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
	assert.Equal(t, fsEvent.Status, event.Status)
	assert.Equal(t, fsEvent.Protocol, event.Protocol)
	assert.Equal(t, fsEvent.SessionID, event.SessionID)
	assert.Equal(t, fsEvent.IP, event.IP)
	assert.Equal(t, fsEvent.FsProvider, event.FsProvider)
	assert.Equal(t, fsEvent.Bucket, event.Bucket)
	assert.Equal(t, fsEvent.Endpoint, event.Endpoint)
	assert.Equal(t, fsEvent.OpenFlags, event.OpenFlags)

	providerEvent := &notifier.ProviderEvent{
		Timestamp:  time.Now().UnixNano(),
		Action:     "add",
		Username:   "adminUsername",
		IP:         "127.0.0.1",
		ObjectType: "admin",
		ObjectName: "adminname",
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
	assert.Equal(t, providerEvent.ObjectData, providerEv.ObjectData)

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
