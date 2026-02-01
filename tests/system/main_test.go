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

	// 1) Setup: create cluster
	_, err = utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: env.Timeouts.CreateCluster},
		"kind", "create", "cluster",
		"--name", env.Cluster.Name,
		"--config", env.Cluster.KindConfig,
	)
	if err != nil {
		// Se quiser idempotência, trate "already exists" aqui (ver seção 4)
		fmt.Fprintln(os.Stderr, "kind create failed:", err)
		os.Exit(1)
	}

	// (Opcional) aplicar namespaces
	// _, _ = utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: env.Timeouts.Apply},
	// 	"kubectl", "--context", env.Cluster.KubeCtx,
	// 	"apply", "-f", filepath.Join(loaded.RepoRoot, "tests/infra/k8s/namespaces.yaml"),
	// )

	// 2) Run tests
	code := m.Run()

	// 3) Teardown (best effort)
	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", env.Cluster.Name)

	os.Exit(code)
}
