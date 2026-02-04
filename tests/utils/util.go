package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CmdOptions struct {
	Dir     string            // working directory
	Env     map[string]string // env overrides (merge com os.Environ)
	Stdin   string            // opcional
	Timeout time.Duration     // se > 0, cria um contexto com timeout
}

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Exec(ctx context.Context, name string, args ...string) error {
	_, err := ExecWithResult(ctx, CmdOptions{}, name, args...)
	return err
}

func ExecWithResult(ctx context.Context, opt CmdOptions, name string, args ...string) (CmdResult, error) {
	if name == "" {
		return CmdResult{}, errors.New("command name is empty")
	}

	// Timeout opcional
	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if opt.Dir != "" {
		cmd.Dir = opt.Dir
	}

	// Env: merge os.Environ + overrides
	if len(opt.Env) > 0 {
		env := append([]string{}, os.Environ()...)
		for k, v := range opt.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if opt.Stdin != "" {
		cmd.Stdin = strings.NewReader(opt.Stdin)
	}

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		// Extrair exit code quando possível
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}

	res := CmdResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	if err != nil {
		return res, fmt.Errorf("cmd failed: %s %s (exit=%d): %w\nstderr:\n%s\nstdout:\n%s",
			name, strings.Join(args, " "), res.ExitCode, err, res.Stderr, res.Stdout)
	}

	return res, nil
}

// RepoRoot returns the absolute path of the git repository root.
// It walks up from the current working directory until it finds ".git".
func RepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("repo root not found: no .git directory")
}
