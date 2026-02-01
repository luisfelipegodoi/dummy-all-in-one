package system

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"tests/config"
	"tests/utils"
	"time"
)

const (
	testsWorkingDirectory = "tests"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfgPath, err := config.Setup("")
	if err != nil {
		panic(err)
	}

	env, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	_ = cfgPath // se quiser logar

	// 1. Setup
	rootFolder, err := utils.RepoRoot()
	if err != nil {
		fmt.Errorf("error to get root folder. details: %s", err.Error())
	}

	clusterManifest := filepath.Join(rootFolder, testsWorkingDirectory, env.ClusterManifest)

	_, err = utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: 2 * time.Minute},
		"kind", "create", "cluster",
		"--config", clusterManifest,
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	// utils.Exec(ctx, "kubectl", "apply", "-f", "tests/infra/k8s/namespaces.yaml")

	// (aqui depois entra ingress, deps, etc)

	// 2. Run tests
	code := m.Run()

	// 3. Teardown
	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", env.ClusterName)

	os.Exit(code)
}
