package store_test

import (
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/smartcontractkit/chainlink/internal/mocks"
	"github.com/smartcontractkit/chainlink/store"
	"github.com/smartcontractkit/chainlink/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Start(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	store := app.Store
	ctrl := gomock.NewController(t)
	txmMock := mocks.NewMockTxManager(ctrl)
	store.TxManager = txmMock
	txmMock.EXPECT().Register(gomock.Any())
	assert.NoError(t, store.Start())
}

func TestStore_Close(t *testing.T) {
	t.Parallel()

	s, cleanup := cltest.NewStore()
	defer cleanup()

	s.RunChannel.Send("whatever")
	s.RunChannel.Send("whatever")

	_, open := <-s.RunChannel.Receive()
	assert.True(t, open)

	_, open = <-s.RunChannel.Receive()
	assert.True(t, open)

	assert.NoError(t, s.Close())

	rr, open := <-s.RunChannel.Receive()
	assert.Equal(t, store.RunRequest{}, rr)
	assert.False(t, open)
}

func TestStore_SyncDiskKeyStoreToDb_HappyPath(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplication()
	defer cleanup()
	store := app.GetStore()

	// create key on disk
	pwd := "p@ssword"
	acc, err := store.KeyStore.NewAccount(pwd)
	require.NoError(t, err)

	// assert creation on disk is successful
	files, err := utils.FilesInDir(app.Config.KeysDir())
	require.NoError(t, err)
	require.Len(t, files, 1)

	// sync
	require.NoError(t, store.SyncDiskKeyStoreToDb())

	// assert creation in db is successful
	keys, err := store.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
	key := keys[0]
	require.Equal(t, acc.Address.Hex(), key.Address.String())

	// assert contents are the same
	content, err := utils.FileContents(filepath.Join(app.Config.KeysDir(), files[0]))
	require.NoError(t, err)
	require.Equal(t, keys[0].JSON.String(), content)
}

func TestStore_SyncDiskKeyStoreToDb_DbKeyAlreadyExists(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	require.NoError(t, app.StartAndConnect())
	store := app.GetStore()

	// assert sync worked on NewApplication
	keys, err := store.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1, "key should already exist because of Application#Start")

	// get account
	acc, err := store.KeyStore.GetFirstAccount()
	require.NoError(t, err)

	require.NoError(t, store.SyncDiskKeyStoreToDb()) // sync

	// assert no change in db
	keys, err = store.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Equal(t, acc.Address.Hex(), keys[0].Address.String())
}

func TestQueuedRunChannel_Send(t *testing.T) {
	t.Parallel()

	rq := store.NewQueuedRunChannel()

	assert.NoError(t, rq.Send("first"))
	rr1 := <-rq.Receive()
	assert.NotNil(t, rr1)
}

func TestQueuedRunChannel_Send_afterClose(t *testing.T) {
	t.Parallel()

	rq := store.NewQueuedRunChannel()
	rq.Close()

	assert.Error(t, rq.Send("first"))
}
