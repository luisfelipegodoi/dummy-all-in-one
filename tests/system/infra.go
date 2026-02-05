package system

import (
	"os"
	"strings"
	"tests/system/spec"
)

func resolvePlan() spec.Plan {
	// 1) prioridade: variável de ambiente
	if flow := os.Getenv("FLOW"); flow != "" {
		return resolveFromFlow(flow)
	}

	// 2) fallback: inferir pelo diretório
	return resolvePlanFromCWD()
}

func resolveFromFlow(flow string) spec.Plan {
	switch flow {

	case "aws_only":
		return spec.Plan{
			"cluster-a": {Localstack: true, DynamoSeed: true},
			"cluster-b": {NATS: false, Redis: false},
		}

	case "event_flow":
		return spec.Plan{
			"cluster-a": {Localstack: true, DynamoSeed: true},
			"cluster-b": {NATS: true, Redis: true},
		}

	case "platform_flow":
		return spec.Plan{
			"cluster-b": {NATS: true, Redis: true, ArgoCD: true},
		}

	default:
		panic("FLOW inválido: " + flow)
	}
}

func resolvePlanFromCWD() spec.Plan {
	wd, _ := os.Getwd()

	switch {
	case strings.Contains(wd, "aws_only"):
		return resolveFromFlow("aws_only")

	case strings.Contains(wd, "event_flow"):
		return resolveFromFlow("event_flow")

	case strings.Contains(wd, "platform_flow"):
		return resolveFromFlow("platform_flow")

	default:
		return resolveFromFlow("aws_only")
	}
}
