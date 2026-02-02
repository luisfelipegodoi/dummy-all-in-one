package system

import (
	"os"
	"strings"
	"tests/system/spec"
)

func resolvePlanFromCWD() spec.Plan {
	wd, _ := os.Getwd()

	switch {
	case strings.Contains(wd, "aws_only"):
		return spec.Plan{
			"cluster-a": {Localstack: true, DynamoSeed: true},
		}

	case strings.Contains(wd, "event_flow"):
		return spec.Plan{
			"cluster-a": {Localstack: true, DynamoSeed: true},
			"cluster-b": {NATS: true, Redis: true},
		}

	case strings.Contains(wd, "platform_flow"):
		return spec.Plan{
			"cluster-b": {NATS: true, Redis: true, ArgoCD: true},
		}

	default:
		//panic("cannot resolve plan: unknown flow folder")
		return spec.Plan{
			"cluster-a": {Localstack: true, DynamoSeed: true},
			"cluster-b": {NATS: true, Redis: true},
		}
	}
}
