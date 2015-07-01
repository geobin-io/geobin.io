package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

const (
	// Path to the config file
	configFile = "./config.json"
)

// Config holds configuration values read in from the config file
type Config struct {
	Host       string
	Port       int
	RedisHost  string
	RedisPass  string
	RedisDB    int64
	NameVals   string
	NameLength int
	RateLimit  int
	BinExpires time.Duration
}

// loadConfig reads configuration values from the config file
func loadConfig() *Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)

	var conf Config
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	conf.BinExpires = conf.BinExpires * time.Hour

	return &conf
}
