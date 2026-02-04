/*
 	Execute this command to add the deleted_at table in postgreSQL
    alter table todos
    add column deleted_at TIMESTAMP NULL
*/

package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

/*
		DeletionMode defines how data is deleted from the database

	 	SOFT - mark as deleted (flag as "deleted_at")
	  	HARD - the row is permanently removed
*/
type DeletionMode string

const (
	DeletionSoft DeletionMode = "SOFT"
	DeletionHard DeletionMode = "HARD"
)

/*
AppConfig holds all application-level configuration values.

Values are loaded from `config.yaml`.
*/
type AppConfig struct {
	DeletionMode DeletionMode `yaml:"deletion_mode"`
}

// Config is the globally accessible configuration instance.

var Config AppConfig

//validDeletionModes defines the allowed enum-like values for the `deletion_mode` configuration.

var validDeletionModes = map[DeletionMode]bool{
	DeletionSoft: true,
	DeletionHard: true,
}

/*
LoadConfig loads configuration from `config.yaml`.

Behavior description:
- Defaults to SOFT delete
- If the file is missing or invalid, defaults are used
- If `deletion_mode` is invalid, it is reset to SOFT
- `Loaded` is set to true only when config is valid
*/
func LoadConfig() {
	// Default configuration (safe behavior)
	Config = AppConfig{
		DeletionMode: DeletionSoft,
	}

	// Attempt to read config.yaml
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("config.yaml not found, defaulting to SOFT delete")
		return
	}

	// Parse YAML into Config
	if err := yaml.Unmarshal(file, &Config); err != nil {
		log.Println("Invalid config.yaml, defaulting to SOFT delete")
		return
	}

	// Validate deletion_mode value
	if !validDeletionModes[Config.DeletionMode] {
		log.Println("Invalid deletion_mode value, defaulting to SOFT delete")
		Config.DeletionMode = DeletionSoft
		return
	}
}
