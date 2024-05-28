package parser

import (
	"syscall/js"
)

type Redis struct {
	client js.Value
}

func NewRedis(client js.Value) *Redis {
	return &Redis{
		client: client,
	}
}

func (r *Redis) Get(key string) (string, error) {
	return r.client.Call("get", key).String(), nil
}

func (r *Redis) Set(key, value string) error {
	r.client.Call("set", key, value)
	return nil
}

func (r *Redis) Del(key string) error {
	r.client.Call("del", key)
	return nil
}

func (r *Redis) Exists(key string) (bool, error) {
	return r.client.Call("exists", key).Bool(), nil
}

func (r *Redis) Ping() error {
	r.client.Call("ping")
	return nil
}
