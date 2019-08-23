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

	return os.ExpandEnv(filepath.Join("$HOME", "noah"))
}

func GetNoahConfigPath() string {
	if NoahConfig != "" {
		return NoahConfig
	}

	return os.ExpandEnv(filepath.Join(GetNoahHome(), "/config/config.toml"))
}
