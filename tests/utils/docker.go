package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Docker struct {
	Timeout time.Duration
}

func (d Docker) Pull(ctx context.Context, image string) error {
	if strings.TrimSpace(image) == "" {
		return errors.New("image is empty")
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"docker", "pull", image,
	)
	return err
}

func (d Docker) ImageExistsLocal(ctx context.Context, image string) (bool, error) {
	if strings.TrimSpace(image) == "" {
		return false, errors.New("image is empty")
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"docker", "image", "inspect", image,
	)
	if err != nil {
		// inspect retorna erro se não existe localmente
		return false, nil
	}
	return true, nil
}

// EnsurePulled pulls only if missing locally.
func (d Docker) EnsurePulled(ctx context.Context, image string) error {
	ok, err := d.ImageExistsLocal(ctx, image)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return d.Pull(ctx, image)
}

// LoadIntoKind loads an already pulled image into a kind cluster.
func (d Docker) LoadIntoKind(ctx context.Context, kindClusterName, image string) error {
	if strings.TrimSpace(kindClusterName) == "" {
		return errors.New("kind cluster name is empty")
	}
	if strings.TrimSpace(image) == "" {
		return errors.New("image is empty")
	}

	timeout := d.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	_, err := ExecWithResult(ctx, CmdOptions{Timeout: timeout},
		"kind", "load", "docker-image", image,
		"--name", kindClusterName,
	)
	if err != nil {
		return fmt.Errorf("kind load docker-image failed: %w", err)
	}
	return nil
}
