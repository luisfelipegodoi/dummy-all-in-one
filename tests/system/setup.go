package system

import (
	"context"
	"tests/config"
	"tests/utils"
)

func SetupInfra(ctx context.Context, spec InfraSpec, env config.Env, loaded config.Loaded) error {
	kub := utils.Kubectl{Context: env.Cluster.KubeCtx}

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
