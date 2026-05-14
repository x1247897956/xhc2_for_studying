package generate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"xhc2_for_studying/protocol"
)

//go:embed implant_src.tar.gz
var implantSrcArchive []byte

//go:embed implant_config.gotmpl
var configTemplateBytes []byte

type implantConfigData struct {
	ServerURL       string
	Interval        int64
	Jitter          int64
	ServerPublicKey string
	C2Profile       *protocol.C2Profile
}

// GenerateImplantConfig 使用 C2Profile 和操作员参数渲染 implant.json。
func GenerateImplantConfig(serverURL string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile) ([]byte, error) {
	tmpl, err := template.New("implant.json").Parse(string(configTemplateBytes))
	if err != nil {
		return nil, fmt.Errorf("parse config template: %w", err)
	}

	data := implantConfigData{
		ServerURL:       serverURL,
		Interval:        interval,
		Jitter:          jitter,
		ServerPublicKey: serverPublicKey,
		C2Profile:       c2Profile,
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
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("write implant config: %w", err)
	}
	return nil
}

// ExtractImplantSource 将嵌入的 implant 源码解压到 destDir。
// destDir 下会创建 implant/ 子目录。
func ExtractImplantSource(destDir string) error {
	gr, err := gzip.NewReader(bytes.NewReader(implantSrcArchive))
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		path := filepath.Join(destDir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", path, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", path, err)
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", path, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("write file %s: %w", path, err)
			}
			f.Close()
		}
	}
	return nil
}
