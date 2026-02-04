package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	Clusters map[string]struct {
		Name       string `mapstructure:"name"`
		KubeCtx    string `mapstructure:"kubeContext"`
		KindConfig string `mapstructure:"kindConfig"`
	} `mapstructure:"clusters"`

	HelmApps map[string]struct {
		Chart     string `mapstructure:"chart"`
		Release   string `mapstructure:"release"`
		Namespace string `mapstructure:"namespace"`
	} `mapstructure:"helm"`

	Timeouts struct {
		CreateCluster time.Duration `mapstructure:"createCluster"`
		Apply         time.Duration `mapstructure:"apply"`
		Helm          time.Duration `mapstructure:"helm"`
	} `mapstructure:"timeouts"`
}

// LoadEnv reads config into Env, applies defaults and validates.
func LoadEnv(v *viper.Viper, repoRoot string) (Env, error) {
	var e Env

	// Defaults
	v.SetDefault("cluster.name", "cluster-a")
	v.SetDefault("timeouts.createCluster", "2m")
	v.SetDefault("timeouts.apply", "2m")

	// Unmarshal
	if err := v.Unmarshal(&e); err != nil {
		return Env{}, fmt.Errorf("unmarshal env: %w", err)
	}

	if err := validateEnv(e); err != nil {
		return Env{}, err
	}
	return e, nil
}

func validateEnv(e Env) error {
	if e.Timeouts.CreateCluster <= 0 {
		return fmt.Errorf("timeouts.createCluster must be > 0 (got %s)", e.Timeouts.CreateCluster)
	}
	if e.Timeouts.Apply <= 0 {
		return fmt.Errorf("timeouts.apply must be > 0 (got %s)", e.Timeouts.Apply)
	}
	return nil
}
