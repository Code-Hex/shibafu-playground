package main

import "github.com/kelseyhightower/envconfig"

type EnvConfig struct {
	GAEInstance string `envconfig:"GAE_INSTANCE"`
	Port        int    `envconfig:"PORT" default:"8080"`
	HostName string `envconfig:"HOST_NAME" default:"play.shibafu.org"`
}

func NewEnvConfig() (*EnvConfig, error) {
	var ec EnvConfig
	err := envconfig.Process("", &ec)
	if err != nil {
		return nil, err
	}
	return &ec, nil
}
