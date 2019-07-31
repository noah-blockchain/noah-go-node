package utils

import (
	"os"
	"path/filepath"
)

var (
	NoaxHome   string
	NoaxConfig string
)

func GetNoaxHome() string {
	if NoaxHome != "" {
		return NoaxHome
	}

	home := os.Getenv("NOAXHOME")

	if home != "" {
		return home
	}

	return os.ExpandEnv(filepath.Join("$HOME", ".noah"))
}

func GetNoaxConfigPath() string {
	if NoaxConfig != "" {
		return NoaxConfig
	}

	return NoaxConfig() + "/config/config.toml"
}
