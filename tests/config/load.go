package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Loaded struct {
	RepoRoot   string
	ConfigPath string
	Viper      *viper.Viper
	Env        Env
}

func Load() (Loaded, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Loaded{}, fmt.Errorf("getwd: %w", err)
	}

	repoRoot, err := findRepoRoot(wd)
	if err != nil {
		return Loaded{}, fmt.Errorf("find repo root: %w", err)
	}

	loader := NewLoader()
	v, cfgPath, err := loader.LoadViper()
	if err != nil {
		return Loaded{}, err
	}

	env, err := LoadEnv(v, repoRoot)
	if err != nil {
		return Loaded{}, err
	}

	return Loaded{
		RepoRoot:   filepath.Clean(repoRoot),
		ConfigPath: cfgPath,
		Viper:      v,
		Env:        env,
	}, nil
}
