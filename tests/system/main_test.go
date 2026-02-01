package system

import (
	"context"
	"fmt"
	"os"
	"testing"
	"tests/utils"
	"time"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Setup
	clusterManifest := fmt.Sprintf("%s/%s", utils.RepoRoot())
	utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: 2 * time.Minute},
		"kind", "create", "cluster",
		"--config", "tests/infra/kind/cluster-a.yaml",
	)

	// utils.Exec(ctx, "kubectl", "apply", "-f", "tests/infra/k8s/namespaces.yaml")

	// (aqui depois entra ingress, deps, etc)

	// 2. Run tests
	code := m.Run()

	// 3. Teardown
	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", "cluster-a")

	os.Exit(code)
}
