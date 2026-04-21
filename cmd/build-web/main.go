package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/pkg/errors"
)

const defaultPNPMVersion = "10.15.0"

func main() {
	ctx := context.Background()
	if err := buildAndExportFrontend(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildAndExportFrontend(ctx context.Context) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	uiDir := filepath.Join(repoRoot, "ui")
	embedDir := filepath.Join(repoRoot, "pkg", "smailnaild", "web", "embed", "public")
	if err := os.RemoveAll(embedDir); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "remove old embed assets")
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return errors.Wrap(err, "connect dagger")
	}
	defer func() { _ = client.Close() }()

	pnpmVersion := strings.TrimSpace(os.Getenv("WEB_PNPM_VERSION"))
	if pnpmVersion == "" {
		pnpmVersion = defaultPNPMVersion
	}

	source := client.Host().Directory(uiDir, dagger.HostDirectoryOpts{
		Exclude: []string{
			"dist",
			"node_modules",
			"storybook-static",
			"tsconfig.tsbuildinfo",
		},
	})

	pathValue := "/pnpm:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	pnpmStore := client.CacheVolume("smailnail-ui-pnpm-store")

	container := client.Container().
		From("node:22-bookworm").
		WithEnvVariable("PNPM_HOME", "/pnpm").
		WithEnvVariable("PATH", pathValue).
		WithMountedCache("/pnpm/store", pnpmStore).
		WithDirectory("/src/ui", source).
		WithWorkdir("/src/ui").
		WithExec([]string{"sh", "-lc", "corepack enable && corepack prepare pnpm@" + pnpmVersion + " --activate"}).
		WithExec([]string{"pnpm", "install", "--frozen-lockfile"}).
		WithExec([]string{"pnpm", "run", "build"})

	if _, err := container.Directory("/src/ui/dist/public").Export(ctx, embedDir); err != nil {
		return errors.Wrap(err, "export built frontend into embed/public")
	}

	gitkeepPath := filepath.Join(embedDir, ".gitkeep")
	if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
		return errors.Wrap(err, "create .gitkeep in embed/public")
	}

	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "get working directory")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
