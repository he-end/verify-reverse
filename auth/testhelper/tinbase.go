package testhelper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func projectRoot() (string, error) {
	if root := os.Getenv("TINBASE_PROJECT_ROOT"); root != "" {
		return root, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found; set TINBASE_PROJECT_ROOT env var")
		}
		dir = parent
	}
}

func StartPGContainer(ctx context.Context) (cleanup func(), err error) {
	root, err := projectRoot()
	if err != nil {
		return func() {}, err
	}

	upCmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d", "--wait")
	upCmd.Dir = root
	upCmd.Stdout = os.Stdout
	upCmd.Stderr = os.Stderr
	if err := upCmd.Run(); err != nil {
		return func() {}, fmt.Errorf("docker compose up failed: %w", err)
	}

	cleanup = func() {
		downCmd := exec.Command("docker", "compose", "down")
		downCmd.Dir = root
		downCmd.Run()
	}

	return cleanup, nil
}
