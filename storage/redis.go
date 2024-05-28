package storage

import (
	utils "github.com/globe-and-citizen/layer8-utils"
)

type redisStorage struct {
	client DBClient
}

func (r *redisStorage) WithOptions(options *StorageOptions) {
	// create a new redis client
	if options.Client == nil {
		panic("StorageOptions.Client is required")
	}
	r.client = options.Client

	// ping the redis server to check if it's alive
	err := r.client.Ping()
	if err != nil {
		panic(err)
	}
}

func (r *redisStorage) AddKey(key string, value *utils.JWK) error {
	// first, convert the JWK to a base64 string
	b64, err := value.ExportAsBase64()
	if err != nil {
		return err
	}

	err = r.client.Set("key:"+key, b64)
	if err != nil {
		return err
	}

	return nil
}

func (r *redisStorage) GetKey(key string) *utils.JWK {
	b64, err := r.client.Get("key:" + key)
	if err != nil {
		return nil
	}

	jwk, err := utils.B64ToJWK(b64)
	if err != nil {
		return nil
	}

	return jwk
}

func (r *redisStorage) AddJWT(key string, value string) error {
	err := r.client.Set("jwt:"+key, value)
	if err != nil {
		return err
	}

	return nil
}

func (r *redisStorage) GetJWT(key string) string {
	jwt, err := r.client.Get("jwt:" + key)
	if err != nil {
		return ""
	}

	return jwt
}

var (
	// inMemStorageInstance is the instance of the in-memory storage
	redisStorageInstance *redisStorage
)

func initRedisStorage() {
	redisStorageInstance = &redisStorage{}
}
