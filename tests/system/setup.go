package system

import (
	"context"
	"fmt"
	"tests/config"
	"tests/system/spec"
	"tests/utils"
)

type ClusterTarget struct {
	Key        string
	Name       string
	KubeCtx    string
	KindConfig string
}

func SetupInfra(ctx context.Context, target ClusterTarget, infra spec.InfraSpec, env config.Env, loaded config.Loaded) error {
	// kind load --name target.Name
	// kubectl --context target.KubeCtx
	// helm --kube-context target.KubeCtx

	if infra.Localstack {
		hm := utils.Helm{
			KubeContext: target.KubeCtx,
			Timeout:     env.Timeouts.Apply,
		}

		opts := utils.HelmInstallOpts{
			Release:   env.HelmApps["localstack"].Release,
			Chart:     fmt.Sprintf("%s/%s", loaded.RepoRoot, env.HelmApps["localstack"].Chart),
			Namespace: env.HelmApps["localstack"].Namespace,
			Wait:      true,
			CreateNS:  true,
		}

		if err := hm.UpgradeInstall(ctx, opts); err != nil {
			fmt.Errorf("error to install helm chart")
		}
	}

	if infra.DynamoSeed {
		// TODO: mover aqui o job do dynamodb
	}

	if infra.NATS {
		// TODO: docker pull + kind load + kubectl apply nats.yaml
	}

	if infra.Redis {
		// TODO: futuro
	}

	if infra.ArgoCD {
		// TODO: futuro
	}

	return nil
}

func TargetsFromEnv(env config.Env) ([]ClusterTarget, error) {
	var out []ClusterTarget
	for key, c := range env.Clusters {
		out = append(out, ClusterTarget{
			Key:        key,
			Name:       c.Name,
			KubeCtx:    c.KubeCtx,
			KindConfig: c.KindConfig,
		})
	}

	return out, nil
}
