package utils

import (
	"os"
	"path/filepath"
)

var (
	NoahHome   string
	NoahConfig string
)

func GetNoahHome() string {
	if NoahHome != "" {
		return NoahHome
	}

	home := os.Getenv("NOAHHOME")

	if home != "" {
		return home
	}

	return os.ExpandEnv(filepath.Join("$HOME", ".noah"))
}

func GetNoahConfigPath() string {
	if NoahConfig != "" {
		return NoahConfig
	}

	return GetNoahHome() + "/config/config.toml"
}
