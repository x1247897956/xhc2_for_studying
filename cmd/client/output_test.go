package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultImplantOutputPathUsesCurrentDirectory(t *testing.T) {
	path := defaultImplantOutputPath("", "linux", "amd64")
	want := filepath.Join(".", "implant-linux-amd64")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestDefaultImplantOutputPathUsesExeForWindows(t *testing.T) {
	path := defaultImplantOutputPath("", "windows", "amd64")
	want := filepath.Join(".", "implant-windows-amd64.exe")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestDefaultImplantOutputPathHonorsExplicitPath(t *testing.T) {
	path := defaultImplantOutputPath("/tmp/custom-agent", "linux", "amd64")
	if path != "/tmp/custom-agent" {
		t.Fatalf("path = %q, want /tmp/custom-agent", path)
	}
}

func TestMainHelpIncludesGenerateExampleAndHelpHint(t *testing.T) {
	help := mainHelpText()
	for _, want := range []string{
		"generate [options]",
		"generate -h",
		"generate -server-url http://127.0.0.1:8024 -os linux -arch amd64 -out ./implant-linux",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help text does not contain %q\n%s", want, help)
		}
	}
}

func TestGenerateUsageIncludesAllSupportedFlags(t *testing.T) {
	usage := generateUsageText()
	for _, want := range []string{
		"usage: generate [options]",
		"-server-url string",
		"-path-prefix string",
		"-interval int",
		"-jitter int",
		"-os string",
		"-arch string",
		"-out string",
	} {
		if !strings.Contains(usage, want) {
			t.Fatalf("generate usage does not contain %q\n%s", want, usage)
		}
	}
}
