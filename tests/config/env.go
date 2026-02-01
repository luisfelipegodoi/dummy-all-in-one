package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	ClusterName          string
	ClusterManifest      string
	KubeContext          string
	CreateClusterTimeout time.Duration
	ApplyTimeout         time.Duration
}

func LoadEnv() (Env, error) {
	bindEnvVars()

	e := Env{
		ClusterName:          viper.GetString("cluster.name"),
		ClusterManifest:      viper.GetString("cluster.manifest"),
		KubeContext:          viper.GetString("cluster.kubeContext"),
		CreateClusterTimeout: viper.GetDuration("timeouts.createCluster"),
		ApplyTimeout:         viper.GetDuration("timeouts.apply"),
	}

	applyDefaults(&e)

	if err := e.Validate(); err != nil {
		return Env{}, err
	}

	return e, nil
}

func bindEnvVars() {
	_ = viper.BindEnv("cluster.name", "DUMMY_CLUSTER_NAME")
	_ = viper.BindEnv("cluster.kindConfig", "DUMMY_CLUSTER_KIND_CONFIG")
	_ = viper.BindEnv("cluster.kubeContext", "DUMMY_CLUSTER_KUBE_CONTEXT")
	_ = viper.BindEnv("timeouts.createCluster", "DUMMY_TIMEOUT_CREATE_CLUSTER")
	_ = viper.BindEnv("timeouts.apply", "DUMMY_TIMEOUT_APPLY")
}

func applyDefaults(e *Env) {
	if strings.TrimSpace(e.ClusterName) == "" {
		e.ClusterName = "system-tests-lab"
	}
	if strings.TrimSpace(e.KubeContext) == "" {
		e.KubeContext = "kind-" + e.ClusterName
	}
	if e.CreateClusterTimeout == 0 {
		e.CreateClusterTimeout = 2 * time.Minute
	}
	if e.ApplyTimeout == 0 {
		e.ApplyTimeout = 2 * time.Minute
	}
}

func (e Env) Validate() error {
	if strings.TrimSpace(e.ClusterName) == "" {
		return errors.New("cluster.name is required")
	}
	if strings.TrimSpace(e.ClusterManifest) == "" {
		return errors.New("cluster.manifest is required")
	}
	if strings.TrimSpace(e.KubeContext) == "" {
		return errors.New("cluster.kubeContext is required")
	}
	if e.CreateClusterTimeout <= 0 {
		return fmt.Errorf("timeouts.createCluster must be > 0 (got %s)", e.CreateClusterTimeout)
	}
	if e.ApplyTimeout <= 0 {
		return fmt.Errorf("timeouts.apply must be > 0 (got %s)", e.ApplyTimeout)
	}
	return nil
}
