package main

import "github.com/99designs/keyring"

type ConfigKey string

const (
	ConfigUserSecret ConfigKey = "usersecret"
	ConfigDeviceID   ConfigKey = "deviceid"
)

type Config struct {
	keyring keyring.Keyring
}

func NewConfig() (*Config, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "com.donatstudios.pullaway",
	})

	if err != nil {
		return nil, err
	}

	return &Config{
		keyring: ring,
	}, nil
}

func (c *Config) GetKey(key ConfigKey) (string, error) {
	pass, err := c.keyring.Get(string(key))
	if err == keyring.ErrKeyNotFound {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return string(pass.Data), err
}

func (c *Config) SetKey(key ConfigKey, value string) error {
	return c.keyring.Set(keyring.Item{
		Key:  string(key),
		Data: []byte(value),
	})
}
