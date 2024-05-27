package storage

import (
	utils "github.com/globe-and-citizen/layer8-utils"
)

type (
	ecdh struct {
		pri *utils.JWK
		pub *utils.JWK
	}
	keys []map[string]*utils.JWK
	jwts []map[string]string
)

func (e *ecdh) GetPrivateKey() *utils.JWK {
	return e.pri
}

func (e *ecdh) GetPublicKey() *utils.JWK {
	return e.pub
}

func (k *keys) Add(key string, value *utils.JWK) {
	*k = append(*k, map[string]*utils.JWK{key: value})
}

func (k *keys) Get(key string) *utils.JWK {
	for _, v := range *k {
		if jwk, ok := v[key]; ok {
			return jwk
		}
	}
	return nil
}

func (j *jwts) Add(key string, value string) {
	*j = append(*j, map[string]string{key: value})
}

func (j *jwts) Get(key string) string {
	for _, v := range *j {
		if jwt, ok := v[key]; ok {
			return jwt
		}
	}
	return ""
}

type inMemStorage struct {
	ECDH *ecdh
	Keys keys
	JWTs jwts
}

var (
	// inMemStorageInstance is the instance of the in-memory storage
	inMemStorageInstance *inMemStorage
)

func InitInMemStorage(pri, pub *utils.JWK) {
	inMemStorageInstance = &inMemStorage{
		ECDH: &ecdh{
			pri: pri,
			pub: pub,
		},
		Keys: []map[string]*utils.JWK{},
		JWTs: []map[string]string{},
	}
}

func GetInMemStorage() *inMemStorage {
	return inMemStorageInstance
}
