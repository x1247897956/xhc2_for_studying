// Package generate handles implant binary compilation and embedded source
// generation.
package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/store"
)

// ImplantBuildResult contains the compiled implant binary and server-side
// metadata recorded during generation.
type ImplantBuildResult struct {
	Digest string
	Binary []byte
}

// BuildImplant compiles the ./implant directory under projectRoot and writes
// the resulting binary to outputPath.
func BuildImplant(projectRoot, outputPath string) error {
	binaryPath := outputPath
	if !filepath.IsAbs(binaryPath) {
		binaryPath = filepath.Join(projectRoot, binaryPath)
	}

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, "./implant")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// GenerateAndBuildEmbedded extracts the embedded implant source, renders a
// configuration using the supplied parameters, cross-compiles for the given
// GOOS/GOARCH, and writes the binary to outputPath. No external project source
// tree is required — only the Go toolchain.
func GenerateAndBuildEmbedded(serverURL, pathPrefix string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile, implantStore store.ImplantStore, outputPath string, goos, goarch string) error {
	result, err := GenerateImplantConfig(serverURL, pathPrefix, interval, jitter, serverPublicKey, c2Profile, implantStore)
	if err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	projectRoot, err := os.MkdirTemp("", "xc2-implant-build-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(projectRoot)

	if err := ExtractImplantSource(projectRoot); err != nil {
		return fmt.Errorf("extract implant source: %w", err)
	}

	configPath := filepath.Join(projectRoot, "implant", "config", "implant.json")
	if err := os.WriteFile(configPath, result.Config, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	env := os.Environ()
	if goos != "" {
		env = append(env, "GOOS="+goos)
	}
	if goarch != "" {
		env = append(env, "GOARCH="+goarch)
	}

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", outputPath, "./implant")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// GenerateAndBuildEmbeddedBytes builds an implant from the embedded source and
// returns the compiled binary instead of writing it to a caller-visible path.
func GenerateAndBuildEmbeddedBytes(serverURL, pathPrefix string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile, implantStore store.ImplantStore, goos, goarch string) (*ImplantBuildResult, error) {
	projectRoot, err := os.MkdirTemp("", "xc2-implant-build-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(projectRoot)

	result, err := GenerateImplantConfig(serverURL, pathPrefix, interval, jitter, serverPublicKey, c2Profile, implantStore)
	if err != nil {
		return nil, fmt.Errorf("generate config: %w", err)
	}

	if err := ExtractImplantSource(projectRoot); err != nil {
		return nil, fmt.Errorf("extract implant source: %w", err)
	}

	configPath := filepath.Join(projectRoot, "implant", "config", "implant.json")
	if err := os.WriteFile(configPath, result.Config, 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	outputPath := filepath.Join(projectRoot, implantFilename(goos, goarch))
	env := os.Environ()
	if goos != "" {
		env = append(env, "GOOS="+goos)
	}
	if goarch != "" {
		env = append(env, "GOARCH="+goarch)
	}

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", outputPath, "./implant")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go build failed: %w", err)
	}

	binary, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read implant binary: %w", err)
	}
	return &ImplantBuildResult{
		Digest: result.Digest,
		Binary: binary,
	}, nil
}

func implantFilename(goos, goarch string) string {
	name := "implant-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name
}
