package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type DeletionMode string

const (
	DeletionSoft DeletionMode = "SOFT"
	DeletionHard DeletionMode = "HARD"
)

type AppConfig struct {
	DeletionMode DeletionMode `yaml:"deletion_mode"`
}

var Config AppConfig

func LoadConfig() {
	Config = AppConfig{
		DeletionMode: DeletionHard,
	}

	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("config.yaml was not found, default configuration will be used")
		return
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		log.Println("Invalid config.yaml, default configuration will be used")
		return
	}

	switch cfg.DeletionMode {
	case DeletionSoft, DeletionHard:
		Config.DeletionMode = cfg.DeletionMode
	default:
		log.Println("Invalid deletion mode, using the default (HARD)")

	}
}
