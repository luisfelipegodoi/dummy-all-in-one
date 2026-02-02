package system

import (
	"context"
	"fmt"
	"os"
	"testing"

	"tests/config"
	"tests/utils"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	loaded, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load failed:", err)
		os.Exit(1)
	}

	env := loaded.Env

	// 1) Create cluster
	if _, err := utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: env.Timeouts.CreateCluster},
		"kind", "create", "cluster",
		"--name", env.Cluster.Name,
		"--config", env.Cluster.KindConfig,
	); err != nil {
		fmt.Fprintln(os.Stderr, "kind create failed:", err)
		os.Exit(1)
	}

	// 2) Resolve infra for this flow
	spec := resolveInfraFromCWD()

	// 3) Setup only what this flow needs
	if err := SetupInfra(ctx, spec, env, loaded); err != nil {
		fmt.Fprintln(os.Stderr, "setup infra failed:", err)
		_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", env.Cluster.Name)
		os.Exit(1)
	}

	// 4) Run tests
	code := m.Run()

	// 5) Teardown
	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", env.Cluster.Name)

	os.Exit(code)
}
