package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestCLIInstallCommand(t *testing.T) {
	t.Run("prints non-mutating install and packaging guidance", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{"install", "--force"})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}
		body := stdout.String()
		for _, want := range []string{
			"No dependencies were installed",
			"go install github.com/TrebuchetDynamics/goscrapling/cmd/goscrapling@latest",
			"Chrome or Chromium",
			"Docker",
			"--force accepted for Scrapling CLI compatibility",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("install output missing %q:\n%s", want, body)
			}
		}
		for _, forbidden := range []string{"playwright install", "apt-get", "brew install"} {
			if strings.Contains(body, forbidden) {
				t.Fatalf("install output suggested live package-manager command %q:\n%s", forbidden, body)
			}
		}
	})

	t.Run("writes deterministic JSON metadata", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(&stdout, &stderr, []string{"install", "--json"})
		if err != nil {
			t.Fatalf("Run returned error: %v\nstderr: %s", err, stderr.String())
		}

		var report struct {
			Command          string   `json:"command"`
			BrowserDownloads bool     `json:"browser_downloads"`
			Runtime          []string `json:"runtime"`
			Docker           []string `json:"docker"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("decode install JSON: %v\n%s", err, stdout.String())
		}
		if report.Command != "goscrapling install" {
			t.Fatalf("command = %q", report.Command)
		}
		if report.BrowserDownloads {
			t.Fatal("install JSON claims browser downloads are performed")
		}
		if !containsInstallText(report.Runtime, "Chrome or Chromium") {
			t.Fatalf("runtime guidance missing browser dependency: %#v", report.Runtime)
		}
		if !containsInstallText(report.Docker, "include Chrome/Chromium") {
			t.Fatalf("docker guidance missing browser dependency: %#v", report.Docker)
		}
	})

	t.Run("help and parse errors stay local", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if err := Run(&stdout, &stderr, []string{"install", "--help"}); err != nil {
			t.Fatalf("install help returned error: %v", err)
		}
		if !strings.Contains(stdout.String(), "usage: goscrapling install") || !strings.Contains(stdout.String(), "does not download browsers") {
			t.Fatalf("unexpected install help:\n%s", stdout.String())
		}

		stdout.Reset()
		stderr.Reset()
		err := Run(&stdout, &stderr, []string{"install", "--download"})
		if !errors.Is(err, ErrParse) {
			t.Fatalf("install --download error = %v, want ErrParse", err)
		}
	})
}

func containsInstallText(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}
