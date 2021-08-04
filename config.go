package main

import (
	"encoding/json"
	"os"
)

var (
	Port = os.Getenv("PORT")
)

type Config struct {
	Port         string
	ClientSecret string
}

func NewConfig() (*Config, error) {
	addr := "8080"
	if Port != "" {
		addr = Port
	}

	b, err := os.Open("secrets.json")
	if err != nil {
		return nil, err
	}
	parsed := make(map[string]string)
	parseErr := json.NewDecoder(b).Decode(&parsed)
	if parseErr != nil {
		return nil, parseErr
	}

	return &Config{
		Port:         addr,
		ClientSecret: parsed["client_secret"],
	}, nil
}
