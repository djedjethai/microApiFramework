package repository

import (
	"errors"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/config"
	mgoCltProvider "gitlab.com/grpasr/common/databases/mongo"
	obs "gitlab.com/grpasr/common/observability"
	// "gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	"sync"
)

const (
	apiServerCollection = "apiServer"
	usersCollection     = "users"
)

type Repository struct {
	ITemporaryStore
	IUserStore
	IAPIserverStore
	IRedisStore
}

func NewRepository(conf *config.Config) *Repository {
	// set configs to create the client
	nonReplicaSetConfig := mgoCltProvider.NewNonReplicaSetConfig(
		conf.MgoGetURL(),
		conf.MgoGetUsername(),
		conf.MgoGetPassword())

	replicaSetConfig := mgoCltProvider.NewReplicaSetConfig(
		conf.MgoGetURL(),
		conf.MgoGetReplicaSetName(),
	)

	storeConfig := mgoCltProvider.NewStoreConfig(
		conf.GlbGetenv(),
		conf.MgoGetUsersDatabaseName(),
		-1,    // connectionTimeout(will default to default)
		-1,    // requestTimeout(will default to default)
		false, // not a replicaSet
		replicaSetConfig,
		nonReplicaSetConfig)

	// create the client
	client, err := mgoCltProvider.ClientProvider(storeConfig, int8(3), int8(5))
	if err != nil {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("NewRepository - create mongo client failed")
		// TODO stop the execution ??
	}

	return &Repository{
		NewTmpStore(),
		NewUserStore(storeConfig, client),
		NewAPIserverStore(storeConfig, client),
		NewRedisStore(conf)}
}

// the LRU or on the app mem
type ITemporaryStore interface {
	TemporarySet(k, d string) error
	TemporaryGet(k string) (string, error)
	TemporaryDelete(k string) error
	TemporaryCount() int
	TemporaryReset()
}

type TmpStore struct {
	str map[string]string
	sync.RWMutex
}

func NewTmpStore() *TmpStore {
	return &TmpStore{
		str: make(map[string]string),
	}
}

func (as *TmpStore) TemporarySet(k, d string) error {
	as.Lock()
	as.str[k] = d
	as.Unlock()
	return nil
}

func (as *TmpStore) TemporaryGet(k string) (string, error) {
	as.RLock()
	dt, ok := as.str[k]
	as.RUnlock()
	if !ok {
		return "", errors.New("Not found")
	}
	return dt, nil
}

func (as *TmpStore) TemporaryDelete(k string) error {
	as.Lock()
	delete(as.str, k)
	as.Unlock()
	return nil
}

func (as *TmpStore) TemporaryCount() int {
	var c int
	as.RLock()
	for range as.str {
		c++
	}
	as.RUnlock()
	return c
}

func (as *TmpStore) TemporaryReset() {
	as.Lock()
	as.str = make(map[string]string)
	as.Unlock()
}
