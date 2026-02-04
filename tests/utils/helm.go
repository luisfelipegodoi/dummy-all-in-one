package utils

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

type Helm struct {
	KubeContext string
	Timeout     time.Duration
}

type HelmInstallOpts struct {
	Release   string
	Chart     string
	Namespace string
	Values    []string
	Set       []string
	Wait      bool
	CreateNS  bool
	ExtraArgs []string
}

func (h Helm) UpgradeInstall(ctx context.Context, opt HelmInstallOpts) error {
	if opt.Release == "" || opt.Chart == "" {
		return fmt.Errorf("helm: Release and Chart are required")
	}
	if opt.Namespace == "" {
		opt.Namespace = "default"
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	args := []string{
		"--kube-context", h.KubeContext,
		"upgrade", "--install", opt.Release, opt.Chart,
		"--namespace", opt.Namespace,
	}

	if opt.CreateNS {
		args = append(args, "--create-namespace")
	}
	if opt.Wait {
		args = append(args, "--wait", "--timeout", timeout.String())
	}

	for _, vf := range opt.Values {
		args = append(args, "-f", filepath.Clean(vf))
	}
	for _, s := range opt.Set {
		args = append(args, "--set", s)
	}
	args = append(args, opt.ExtraArgs...)

	res, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"helm", args...,
	)

	if err != nil {
		fmt.Printf("HELM STDOUT:\n%s\n", res.Stdout)
		fmt.Printf("HELM STDERR:\n%s\n", res.Stderr)
		return err
	}

	return err
}

func (h Helm) DependencyBuild(ctx context.Context, chartPath string) error {
	timeout := h.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"helm",
		"dependency", "build", chartPath,
	)
	return err
}

func (h Helm) Uninstall(ctx context.Context, release, namespace string) error {
	if release == "" {
		return fmt.Errorf("helm: release is required")
	}
	if namespace == "" {
		namespace = "default"
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"helm",
		"--kube-context", h.KubeContext,
		"uninstall", release,
		"--namespace", namespace,
	)
	return err
}
