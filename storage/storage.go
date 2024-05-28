package storage

import utils "github.com/globe-and-citizen/layer8-utils"

const (
	STORAGE_INMEM = iota + 1
	STORAGE_REDIS
)

type ecdh struct {
	pri *utils.JWK
	pub *utils.JWK
}

var ECDH *ecdh

func (e *ecdh) GetPrivateKey() *utils.JWK {
	return e.pri
}

func (e *ecdh) GetPublicKey() *utils.JWK {
	return e.pub
}

type (
	// Storage is the interface for the storage
	Storage interface {
		AddKey(key string, value *utils.JWK) error
		GetKey(key string) *utils.JWK
		AddJWT(key string, value string) error
		GetJWT(key string) string
	}

	// DBClient is a generic interface for database clients
	DBClient interface {
		Get(key string) (string, error)
		Set(key, value string) error
		Del(key string) error
		Exists(key string) (bool, error)
		Ping() error
	}
)

// StorageOptions is the options for the storage
type StorageOptions struct {
	Host     string
	Port     int
	Password string
	DB       int
	// when Client is set, the other fields are ignored
	Client DBClient
}

// GetStorage returns the storage implementation
func GetStorage(t int) Storage {
	switch t {
	case STORAGE_INMEM:
		return inMemStorageInstance
	case STORAGE_REDIS:
		return redisStorageInstance
	default:
		return nil
	}
}

// InitStorage initializes the storage
func InitStorage(pri, pub *utils.JWK) {
	ECDH = &ecdh{
		pri: pri,
		pub: pub,
	}

	initInMemStorage()
	initRedisStorage()
}

func WithOptions(t int, options *StorageOptions) {
	switch t {
	case STORAGE_REDIS:
		redisStorageInstance.WithOptions(options)
	case STORAGE_INMEM:
		inMemStorageInstance.WithOptions(options)
	}
}
