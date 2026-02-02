package system

import (
	"context"
	"fmt"
	"tests/config"
	"tests/utils"
)

type ClusterTarget struct {
	Key        string // cluster-a (chave no env)
	Name       string
	KubeCtx    string
	KindConfig string
}

func SetupInfra(ctx context.Context, spec InfraSpec, env config.Env, loaded config.Loaded) error {
	kub := utils.Kubectl{Context: env.Cluster.KubeCtx}

	fmt.Println(kub)

	if spec.Localstack {
		// TODO: mover aqui o código de helm localstack do main_test
	}

	if spec.DynamoSeed {
		// TODO: mover aqui o job do dynamodb
	}

	if spec.NATS {
		// TODO: docker pull + kind load + kubectl apply nats.yaml
	}

	if spec.Redis {
		// TODO: futuro
	}

	if spec.ArgoCD {
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
	// opcional: ordenar para determinismo
	return out, nil
}
