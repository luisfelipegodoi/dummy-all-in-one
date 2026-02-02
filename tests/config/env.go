package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	Cluster struct {
		Name       string `mapstructure:"name"`
		KindConfig string `mapstructure:"kindConfig"`
		KubeCtx    string `mapstructure:"kubeContext"`
	} `mapstructure:"cluster"`

	Localstack struct {
		Release   string `mapstructure:"release"`
		Namespace string `mapstructure:"namespace"`
	}

	Timeouts struct {
		CreateCluster time.Duration `mapstructure:"createCluster"`
		Apply         time.Duration `mapstructure:"apply"`
	} `mapstructure:"timeouts"`
}

// LoadEnv reads config into Env, applies defaults and validates.
func LoadEnv(v *viper.Viper, repoRoot string) (Env, error) {
	var e Env

	// Defaults
	v.SetDefault("cluster.name", "system-tests-lab")
	v.SetDefault("timeouts.createCluster", "2m")
	v.SetDefault("timeouts.apply", "2m")

	// Unmarshal
	if err := v.Unmarshal(&e); err != nil {
		return Env{}, fmt.Errorf("unmarshal env: %w", err)
	}

	// Defaults derivados
	if strings.TrimSpace(e.Cluster.KubeCtx) == "" {
		e.Cluster.KubeCtx = "kind-" + e.Cluster.Name
	}

	// Resolve path: aceita relativo ao repo root
	if e.Cluster.KindConfig != "" && !filepath.IsAbs(e.Cluster.KindConfig) && repoRoot != "" {
		e.Cluster.KindConfig = filepath.Join(repoRoot, e.Cluster.KindConfig)
	}

	if err := validateEnv(e); err != nil {
		return Env{}, err
	}
	return e, nil
}

func validateEnv(e Env) error {
	if strings.TrimSpace(e.Cluster.Name) == "" {
		return errors.New("cluster.name is required")
	}
	if strings.TrimSpace(e.Cluster.KindConfig) == "" {
		return errors.New("cluster.kindConfig is required")
	}
	if e.Timeouts.CreateCluster <= 0 {
		return fmt.Errorf("timeouts.createCluster must be > 0 (got %s)", e.Timeouts.CreateCluster)
	}
	if e.Timeouts.Apply <= 0 {
		return fmt.Errorf("timeouts.apply must be > 0 (got %s)", e.Timeouts.Apply)
	}
	return nil
}
