package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Loader struct {
	FileName string // default: env.yaml
}

func NewLoader() Loader {
	return Loader{FileName: "env.yaml"}
}

// LoadViper finds env.yaml and returns a configured viper instance (no global state).
func (l Loader) LoadViper() (*viper.Viper, string, error) {
	start, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("getwd: %w", err)
	}

	// Prefer deterministic: repo root + /tests/env.yaml
	if repoRoot, err := findRepoRoot(start); err == nil {
		candidate := filepath.Join(repoRoot, l.FileName)
		if fileExists(candidate) {
			return readViper(candidate)
		}
	}

	// Fallback: walk upwards from cwd until find env.yaml
	path, err := findUpwards(start, l.FileName)
	if err != nil {
		return nil, "", err
	}
	return readViper(path)
}

func readViper(configPath string) (*viper.Viper, string, error) {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("abs(%s): %w", configPath, err)
	}

	v := viper.New()
	v.SetConfigFile(abs)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, "", fmt.Errorf("read config %s: %w", abs, err)
	}
	return v, abs, nil
}

func findUpwards(dir, file string) (string, error) {
	dir = filepath.Clean(dir)
	for {
		c := filepath.Join(dir, file)
		if fileExists(c) {
			return c, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("%s not found walking up from cwd", file)
}

func findRepoRoot(dir string) (string, error) {
	dir = filepath.Clean(dir)
	for {
		if fileExists(filepath.Join(dir, "env.yaml")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("repo root not found (env.yaml)")
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
