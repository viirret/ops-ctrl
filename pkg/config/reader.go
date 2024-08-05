package config

import (
	"log"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Aliases   map[string]string `toml:"aliases"`
	Autostart map[string]string `toml:"autostart"`
}

var (
	config   Config
	loadOnce sync.Once
)

func LoadConfig(filePath string) {
	loadOnce.Do(func() {
		if _, err := toml.DecodeFile(filePath, &config); err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	})
}

func GetConfig() Config {
	return config
}
