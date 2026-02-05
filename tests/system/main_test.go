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
		fmt.Errorf("error to load configs")
		os.Exit(1)
	}

	env := loaded.Env

	// 1) resolver plano (o que vai em qual cluster)
	plan := resolvePlan()

	// 2) criar só os clusters que serão usados nesse flow
	for key, infra := range plan {
		_ = infra // só pra mostrar que usamos plan

		c, ok := env.Clusters[key]
		if !ok {
			fmt.Fprintln(os.Stderr, "cluster not found in env.yaml:", key)
			os.Exit(1)
		}

		if _, err := utils.ExecWithResult(ctx, utils.CmdOptions{Timeout: env.Timeouts.CreateCluster},
			"kind", "create", "cluster",
			"--name", c.Name,
			"--config", fmt.Sprintf("%s/%s", loaded.RepoRoot, c.KindConfig),
		); err != nil {
			fmt.Fprintln(os.Stderr, "kind create failed:", err)
			os.Exit(1)
		}
	}

	// 3) SetupInfra por cluster-alvo
	for key, infra := range plan {
		c := env.Clusters[key]

		target := ClusterTarget{
			Key:        key,
			Name:       c.Name,
			KubeCtx:    c.KubeCtx,
			KindConfig: c.KindConfig,
		}

		if err := SetupInfra(ctx, target, infra, env, loaded); err != nil {
			fmt.Fprintln(os.Stderr, "setup infra failed:", err)

			for key2 := range plan {
				c2 := env.Clusters[key2]
				_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", c2.Name)
			}
			os.Exit(1)
		}
	}

	// 4) Run tests
	code := m.Run()

	// 5) Teardown só dos clusters do plano
	// for key := range plan {
	// 	c := env.Clusters[key]
	// 	_ = utils.Exec(ctx, "kind", "delete", "cluster", "--name", c.Name)
	// }

	os.Exit(code)
}
