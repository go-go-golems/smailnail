//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := buildAndCopy(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildAndCopy() error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	frontendDir := filepath.Join(repoRoot, "ui")
	distDir := filepath.Join(frontendDir, "dist", "public")
	embedDir := filepath.Join(repoRoot, "pkg", "smailnaild", "web", "embed", "public")

	if err := runFrontendBuild(frontendDir, embedDir); err != nil {
		return err
	}

	if _, err := os.Stat(distDir); err != nil {
		if os.IsNotExist(err) && hasGeneratedEmbedAssets(embedDir) {
			fmt.Printf("Skipping embed refresh: %s is missing; keeping committed embed/public assets.\n", distDir)
			return nil
		}
		return fmt.Errorf("stat dist dir: %w", err)
	}

	if err := os.RemoveAll(embedDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove embed dir: %w", err)
	}
	if err := os.MkdirAll(embedDir, 0o755); err != nil {
		return fmt.Errorf("create embed dir: %w", err)
	}

	fmt.Printf("Copying %s to %s...\n", distDir, embedDir)
	if err := copyDir(distDir, embedDir); err != nil {
		return fmt.Errorf("copy dist to embed: %w", err)
	}

	fmt.Println("Frontend build and copy completed successfully!")
	return nil
}

func runFrontendBuild(frontendDir, embedDir string) error {
	if !hasFrontendDependencies(frontendDir) {
		if hasGeneratedEmbedAssets(embedDir) {
			fmt.Println("Skipping frontend rebuild: ui/node_modules is missing; reusing committed embed/public assets.")
			return nil
		}
		return fmt.Errorf("frontend dependencies are missing in %s/ui/node_modules", frontendDir)
	}

	buildTool, toolArgs, err := frontendBuildCommand()
	if err != nil {
		if hasGeneratedEmbedAssets(embedDir) {
			fmt.Printf("Skipping frontend rebuild: %v\n", err)
			return nil
		}
		return err
	}

	fmt.Printf("Building frontend with %s...\n", strings.Join(append([]string{buildTool}, toolArgs...), " "))
	buildCmd := exec.Command(buildTool, toolArgs...)
	buildCmd.Dir = frontendDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("vite build failed: %w", err)
	}
	return nil
}

func frontendBuildCommand() (string, []string, error) {
	if _, err := exec.LookPath("pnpm"); err == nil {
		return "pnpm", []string{"run", "build"}, nil
	}
	if _, err := exec.LookPath("npm"); err == nil {
		return "npm", []string{"run", "build"}, nil
	}
	return "", nil, fmt.Errorf("neither pnpm nor npm is available in PATH")
}

func hasGeneratedEmbedAssets(embedDir string) bool {
	if _, err := os.Stat(filepath.Join(embedDir, "index.html")); err != nil {
		return false
	}
	return true
}

func hasFrontendDependencies(frontendDir string) bool {
	info, err := os.Stat(filepath.Join(frontendDir, "node_modules"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func findRepoRoot() (string, error) {
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
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
