package storage

import (
	utils "github.com/globe-and-citizen/layer8-utils"
)

type inMemStorage struct {
	keys []map[string]*utils.JWK
	jwts []map[string]string
}

func (s *inMemStorage) WithOptions(options *StorageOptions) {
	// no-op
}

func (s *inMemStorage) AddKey(key string, value *utils.JWK) error {
	s.keys = append(s.keys, map[string]*utils.JWK{key: value})
	return nil
}

func (s *inMemStorage) GetKey(key string) *utils.JWK {
	for _, v := range s.keys {
		if jwk, ok := v[key]; ok {
			return jwk
		}
	}
	return nil
}

func (s *inMemStorage) AddJWT(key string, value string) error {
	s.jwts = append(s.jwts, map[string]string{key: value})
	return nil
}

func (s *inMemStorage) GetJWT(key string) string {
	for _, v := range s.jwts {
		if jwt, ok := v[key]; ok {
			return jwt
		}
	}
	return ""
}

var (
	// inMemStorageInstance is the instance of the in-memory storage
	inMemStorageInstance *inMemStorage
)

func initInMemStorage() {
	inMemStorageInstance = &inMemStorage{
		keys: []map[string]*utils.JWK{},
		jwts: []map[string]string{},
	}
}
