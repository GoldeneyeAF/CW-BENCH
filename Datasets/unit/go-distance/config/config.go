package main

import (
	"errors"
	"strconv"
	"time"
)

type Config struct {
	HTTPServer HTTPServer
	Database   DatabaseConfig
}

type HTTPServer struct {
	Port        string
	IdleTimeout time.Duration
}

type DatabaseConfig struct {
	DSN                 string
	MigrationsDirectory string
}

func NewConfig(port string, idleTimeout string, DSN string, MigrationsDirectory string) (*Config, error) {
	numberOfSeconds, err := strconv.Atoi(idleTimeout)
	if err != nil {
		return nil, err
	}
	if numberOfSeconds < 0 {
		err = errors.New("Provided timeout with negative number")
		return nil, err
	}
	numberOfPort, err := strconv.Atoi(port[1:])
	if err != nil {
		return nil, err
	}
	if numberOfPort < 0 {
		err = errors.New("Provided port with negative number")
		return nil, err
	}
	cfg := Config{HTTPServer{port, time.Duration(numberOfSeconds)}, DatabaseConfig{DSN, MigrationsDirectory}}
	return &cfg, nil
}
