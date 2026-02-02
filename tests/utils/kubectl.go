package utils

import (
	"context"
	"fmt"
	"time"
)

type Kubectl struct {
	Context string
	Timeout time.Duration
}

func (k Kubectl) EnsureNamespace(ctx context.Context, name string) error {
	timeout := k.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// kubectl create ns localstack --dry-run=client -o yaml | kubectl apply -f -
	manifest := fmt.Sprintf(
		"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: %s\n",
		name,
	)

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout, Stdin: manifest},
		"kubectl",
		"--context", k.Context,
		"apply", "-f", "-",
	)
	return err
}

func (k Kubectl) ApplyFile(ctx context.Context, path string) error {
	timeout := k.Timeout
	if timeout == 0 {
		timeout = 1 * time.Minute
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"kubectl",
		"--context", k.Context,
		"apply", "-f", path,
	)
	return err
}

func (k Kubectl) WaitDeploymentReady(ctx context.Context, namespace, name string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"kubectl",
		"--context", k.Context,
		"-n", namespace,
		"rollout", "status",
		"deployment/"+name,
		"--timeout", timeout.String(),
	)
	return err
}
