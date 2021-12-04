package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFsEvent(t *testing.T) {
	n := Notifier{
		InstanceID: "sftpgo1",
	}

	fsEvent := FsEvent{
		Timestamp:         time.Now().UnixNano(),
		Action:            "upload",
		Username:          "username",
		FsPath:            "/tmp/file.txt",
		FsTargetPath:      "/tmp/target.txt",
		VirtualPath:       "file.txt",
		VirtualTargetPath: "target.txt",
		SSHCmd:            "scp",
		FileSize:          123,
		Status:            1,
		Protocol:          "SFTP",
		SessionID:         uuid.NewString(),
		IP:                "::1",
	}

	err := n.NotifyFsEvent(fsEvent.Timestamp, fsEvent.Action, fsEvent.Username, fsEvent.FsPath, fsEvent.FsTargetPath,
		fsEvent.SSHCmd, fsEvent.Protocol, fsEvent.IP, fsEvent.VirtualPath, fsEvent.VirtualTargetPath, fsEvent.SessionID,
		fsEvent.FileSize, fsEvent.Status)
	assert.NoError(t, err)

	sess, cancel := GetDefaultSession()
	defer cancel()

	var event FsEvent
	err = sess.First(&event).Error
	assert.NoError(t, err)

	assert.Equal(t, n.InstanceID, event.InstanceID)
	assert.NotEmpty(t, event.ID)

	fsEvent.ID = event.ID
	fsEvent.InstanceID = event.InstanceID
	assert.Equal(t, fsEvent, event)

	providerEvent := ProviderEvent{
		Timestamp:  time.Now().UnixNano(),
		Action:     "add",
		Username:   "adminUsername",
		IP:         "127.0.0.1",
		ObjectType: "admin",
		ObjectName: "adminname",
		ObjectData: []byte("data"),
	}

	err = n.NotifyProviderEvent(providerEvent.Timestamp, providerEvent.Action, providerEvent.Username,
		providerEvent.ObjectType, providerEvent.ObjectName, providerEvent.IP, providerEvent.ObjectData)
	assert.NoError(t, err)

	var providerEv ProviderEvent
	err = sess.First(&providerEv).Error
	assert.NoError(t, err)

	assert.Equal(t, n.InstanceID, providerEv.InstanceID)
	assert.NotEmpty(t, providerEv.ID)

	providerEvent.ID = providerEv.ID
	providerEvent.InstanceID = providerEv.InstanceID
	assert.Equal(t, providerEvent, providerEv)

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
