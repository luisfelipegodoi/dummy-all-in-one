package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Default file name (sem extensão).
const defaultConfigName = "env"

// Setup finds env.yaml walking upwards from startDir (or current working dir if empty),
// configures Viper, and reads the config file.
//
// It returns the absolute path of the config file that was loaded.
func Setup(startDir string) (string, error) {
	if startDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		startDir = wd
	}

	configPath, err := findConfigUpwards(startDir, "env.yaml")
	if err != nil {
		return "", err
	}

	// Reset global viper state (important for tests / multiple runs)
	viper.Reset()

	// Ler um arquivo específico é mais determinístico do que AddConfigPath+Search.
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Opcional: permitir override por variáveis de ambiente
	// Ex: CLUSTER_NAME sobrescreve cluster.name
	viper.SetEnvPrefix("DUMMY") // DUMMY_CLUSTER_NAME, etc (você pode trocar)
	viper.AutomaticEnv()

	// Para env vars baterem com keys com ponto:
	// DUMMY_CLUSTER_NAME -> cluster.name (via BindEnv no env.go)
	// (Mantém explícito e previsível.)

	if err := viper.ReadInConfig(); err != nil {
		return "", fmt.Errorf("read config %s: %w", configPath, err)
	}

	return configPath, nil
}

// findConfigUpwards starts at dir and walks up until it finds fileName.
// Returns the absolute path to the found file.
func findConfigUpwards(dir, fileName string) (string, error) {
	dir = filepath.Clean(dir)

	for {
		candidate := filepath.Join(dir, fileName)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return "", fmt.Errorf("abs path: %w", err)
			}
			return abs, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return "", errors.New("env.yaml not found walking up from start directory")
}
