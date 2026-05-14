package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"xhc2_for_studying/protocol"
)

// BuildImplant 编译 projectRoot 下的 ./implant 目录，产出二进制到 outputPath。
func BuildImplant(projectRoot, outputPath string) error {
	binaryPath := outputPath
	if !filepath.IsAbs(binaryPath) {
		binaryPath = filepath.Join(projectRoot, binaryPath)
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, "./implant")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// GenerateAndBuildEmbedded 使用嵌入的 implant 源码生成并编译。
// 无需项目源码，仅需 Go 工具链。
func GenerateAndBuildEmbedded(serverURL string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile, outputPath string, goos, goarch string) error {
	configBytes, err := GenerateImplantConfig(serverURL, interval, jitter, serverPublicKey, c2Profile)
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
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	env := os.Environ()
	if goos != "" {
		env = append(env, "GOOS="+goos)
	}
	if goarch != "" {
		env = append(env, "GOARCH="+goarch)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, "./implant")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}
