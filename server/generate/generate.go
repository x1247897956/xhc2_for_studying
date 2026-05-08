package generate

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"xhc2_for_studying/protocol"
)

//go:embed implant_config.tmpl.json
var configTemplateBytes []byte

type implantConfigData struct {
	ServerURL string
	Interval  int64
	Jitter    int64
	C2Profile *protocol.C2Profile
}

// GenerateImplantConfig 使用 C2Profile 和操作员参数渲染 implant.json。
func GenerateImplantConfig(serverURL string, interval, jitter int64, c2Profile *protocol.C2Profile) ([]byte, error) {
	tmpl, err := template.New("implant.json").Parse(string(configTemplateBytes))
	if err != nil {
		return nil, fmt.Errorf("parse config template: %w", err)
	}

	data := implantConfigData{
		ServerURL: serverURL,
		Interval:  interval,
		Jitter:    jitter,
		C2Profile: c2Profile,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute config template: %w", err)
	}

	return buf.Bytes(), nil
}

// WriteImplantConfig 将渲染后的配置写入 implant/config/implant.json。
func WriteImplantConfig(projectRoot string, configBytes []byte) error {
	configPath := filepath.Join(projectRoot, "implant", "config", "implant.json")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("write implant config: %w", err)
	}
	return nil
}
