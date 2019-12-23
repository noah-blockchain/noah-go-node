package utils

import (
	"fmt"
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

func GetNoahConfigPath(networkID string) string {
	if NoahConfig != "" {
		return NoahConfig
	}

	return os.ExpandEnv(filepath.Join(GetNoahHome(), fmt.Sprintf("/config-%s/config.toml", networkID)))
}
