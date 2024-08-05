package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Aliases map[string]string `toml:"aliases"`
}

// LoadAliases reads the TOML file and returns the aliases map.
func LoadAliases(filePath string) (map[string]string, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return config.Aliases, nil
}
