package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"series-batch-go/internal/config/db"
	"series-batch-go/internal/config/gemini"
	"series-batch-go/internal/config/log"
)

type Config struct {
	Gemini gemini.Config `json:"gemini"`
	DB     db.Config     `json:"db"`
	Logger log.Config    `json:"logger"`
}

func Read() *Config {
	profile := os.Getenv("profile")
	if profile == "" {
		panic("profile is not set")
	}
	path := filepath.Join("config", "config."+profile+".json")
	return readFile(path)
}

func readFile(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	var cfg Config
	dec := json.NewDecoder(file)
	err = dec.Decode(&cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}
