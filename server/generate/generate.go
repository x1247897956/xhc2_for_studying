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
	"xhc2_for_studying/server/store"
)

//go:embed implant_src.tar.gz
var implantSrcArchive []byte

//go:embed implant_config.gotmpl
var configTemplateBytes []byte

// implantConfigData holds all template variables for implant.json rendering.
type implantConfigData struct {
	ServerURL            string
	PathPrefix           string
	ImplantAgePublicKey  string
	ImplantAgePrivateKey string
	Interval             int64
	Jitter               int64
	ServerPublicKey      string
	C2Profile            *protocol.C2Profile
	ExtMap               protocol.ExtensionMap
}

// ImplantConfigResult contains the rendered implant configuration and the
// digest used by the server to recognize the generated implant.
type ImplantConfigResult struct {
	Config []byte
	Digest string
}

// GenerateImplantConfig renders implant.json from the C2 profile and operator
// parameters. Each call generates a fresh Age key pair and extension map for
// the implant and stores them in the ImplantStore keyed by the public key
// digest.
func GenerateImplantConfig(serverURL, pathPrefix string, interval, jitter int64, serverPublicKey string, c2Profile *protocol.C2Profile, implantStore store.ImplantStore) (*ImplantConfigResult, error) {
	tmpl, err := template.New("implant.json").Parse(string(configTemplateBytes))
	if err != nil {
		return nil, fmt.Errorf("parse config template: %w", err)
	}

	// Generate a fresh Age key pair for this implant.
	privKey, pubKey, err := protocol.GenerateAgeKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate implant age key pair: %w", err)
	}

	extMap := protocol.GenerateExtensionMap(c2Profile)

	// Index the implant record by the public key digest.
	digest := protocol.PubKeyDigest(pubKey)
	if err := implantStore.Set(digest, &store.ImplantRecord{
		ImplantAgePrivateKey: privKey,
		ExtMap:               extMap,
	}); err != nil {
		return nil, err
	}

	data := implantConfigData{
		ServerURL:            serverURL,
		PathPrefix:           pathPrefix,
		ImplantAgePublicKey:  pubKey,
		ImplantAgePrivateKey: privKey,
		Interval:             interval,
		Jitter:               jitter,
		ServerPublicKey:      serverPublicKey,
		C2Profile:            c2Profile,
		ExtMap:               extMap,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute config template: %w", err)
	}

	return &ImplantConfigResult{
		Config: buf.Bytes(),
		Digest: digest,
	}, nil
}

// WriteImplantConfig writes the rendered configuration bytes to
// implant/config/implant.json under projectRoot.
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

// ExtractImplantSource unpacks the embedded implant source archive into
// destDir, creating an implant/ subdirectory underneath.
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
