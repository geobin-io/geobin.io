package main

import (
	"encoding/json"
	"log"
	"os"
)

// Path to the config file
var configFile = "config.json"

// Config holds configuration values read in from the config file
type Config struct {
	Host       string
	Port       int
	RedisHost  string
	RedisPass  string
	RedisDB    int64
	NameVals   string
	NameLength int
}

// loadConfig reads configuration values from the config file
func loadConfig() {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}
