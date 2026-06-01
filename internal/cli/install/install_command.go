package install

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
)

const installUsage = "usage: goscrapling install [--force] [--json]\n\nPrint Go-native installation, browser runtime, and Docker packaging guidance. This command does not download browsers or install system packages."

type installReport struct {
	Command          string   `json:"command"`
	BrowserDownloads bool     `json:"browser_downloads"`
	Force            bool     `json:"force"`
	Install          []string `json:"install"`
	Runtime          []string `json:"runtime"`
	Docker           []string `json:"docker"`
}

func Run(stdout io.Writer, args []string) error {
	var force bool
	var jsonOutput bool
	for _, arg := range args {
		switch arg {
		case "help", "--help", "-h":
			_, err := fmt.Fprintln(stdout, installUsage)
			return err
		case "--force", "-f":
			force = true
		case "--json":
			jsonOutput = true
		default:
			return diagnostics.ParseError("unknown install option %q", arg)
		}
	}

	report := newInstallReport(force)
	if jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	return writeInstallReport(stdout, report)
}

func newInstallReport(force bool) installReport {
	return installReport{
		Command:          "goscrapling install",
		BrowserDownloads: false,
		Force:            force,
		Install: []string{
			"Install the CLI with: go install github.com/TrebuchetDynamics/goscrapling/cmd/goscrapling@latest",
			"Static extraction, parsing, shell -c shortcuts, and spider/fetcher library APIs use Go module dependencies only.",
		},
		Runtime: []string{
			"Browser extract modes use chromedp and require Chrome or Chromium to be available on the host when those modes are run.",
			"goscrapling install intentionally does not download browsers, run package managers, or mutate host state.",
		},
		Docker: []string{
			"No official goscrapling Docker image is published by this repository yet.",
			"For containers, copy the goscrapling binary into the image and include Chrome/Chromium only if browser fetch modes are needed.",
		},
	}
}

func writeInstallReport(stdout io.Writer, report installReport) error {
	if _, err := fmt.Fprintln(stdout, "goscrapling install"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "No dependencies were installed."); err != nil {
		return err
	}
	if report.Force {
		if _, err := fmt.Fprintln(stdout, "--force accepted for Scrapling CLI compatibility; it remains a no-op in goscrapling."); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(stdout); err != nil {
		return err
	}
	sections := []struct {
		title string
		items []string
	}{
		{title: "Install", items: report.Install},
		{title: "Browser runtime", items: report.Runtime},
		{title: "Docker", items: report.Docker},
	}
	for _, section := range sections {
		if _, err := fmt.Fprintf(stdout, "%s:\n", section.title); err != nil {
			return err
		}
		for _, item := range section.items {
			if _, err := fmt.Fprintf(stdout, "  - %s\n", item); err != nil {
				return err
			}
		}
	}
	return nil
}
