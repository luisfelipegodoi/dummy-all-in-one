package system

import (
	"os"
	"strings"
	"tests/system/flows/aws_only"
	"tests/system/flows/event_flow"
	"tests/system/flows/platform_flow"
)

func resolveInfraFromCWD() InfraSpec {
	wd, _ := os.Getwd()

	switch {
	case strings.Contains(wd, "aws_only"):
		return aws_only.Infra
	case strings.Contains(wd, "event_flow"):
		return event_flow.Infra
	case strings.Contains(wd, "platform_flow"):
		return platform_flow.Infra
	default:
		panic("cannot resolve infra: unknown flow folder")
	}
}
