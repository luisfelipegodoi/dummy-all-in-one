package system

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		fmt.Fprintln(os.Stderr, "kind create failed:", err)
		os.Exit(1)
	}

	chartPath := filepath.Join(loaded.RepoRoot, "infra", "helm", "charts", "localstack")

	helm := utils.Helm{KubeContext: env.Cluster.KubeCtx, Timeout: 5 * time.Minute}

	if err := helm.DependencyBuild(ctx, chartPath); err != nil {
		fmt.Fprintln(os.Stderr, "helm dependency build failed:", err)
		os.Exit(1)
	}

	if err = helm.UpgradeInstall(ctx, utils.HelmInstallOpts{
		Release:   env.Localstack.Release,
		Chart:     chartPath,
		Namespace: env.Localstack.Namespace,
		CreateNS:  true,
		Wait:      true,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	jobPath := filepath.Join(loaded.RepoRoot, "infra", "k8s", "localstack-dynamodb-tables-job.yaml")

	kub := utils.Kubectl{
		Context: env.Cluster.KubeCtx,
		Timeout: 2 * time.Minute,
	}

	// (optional) recriar o job pra rodar sempre
	_, _ = utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: 30 * time.Second},
		"kubectl", "--context", kub.Context,
		"-n", "localstack",
		"delete", "job", "localstack-dynamodb-init",
		"--ignore-not-found=true",
	)

	// aplica o job
	if err := kub.ApplyFile(ctx, jobPath); err != nil {
		fmt.Fprintln(os.Stderr, "apply job failed:", err)
		os.Exit(1)
	}

	// espera completar
	waitTimeout := 3 * time.Minute
	if _, err := utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: waitTimeout},
		"kubectl", "--context", kub.Context,
		"-n", "localstack",
		"wait", "--for=condition=complete",
		"job/localstack-dynamodb-init",
		"--timeout", waitTimeout.String(),
	); err != nil {

		// logs para debug (best effort)
		_, _ = utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: 30 * time.Second},
			"kubectl", "--context", kub.Context,
			"-n", "localstack",
			"logs", "job/localstack-dynamodb-init",
		)

		fmt.Fprintln(os.Stderr, "job failed:", err)
		os.Exit(1)
	}

	// 2) Run tests
	code := m.Run()

	// 3) Teardown (best effort)
	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", env.Cluster.Name)

	os.Exit(code)
}
