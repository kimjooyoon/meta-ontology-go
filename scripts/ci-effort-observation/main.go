package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	config := parseConfig()
	report, err := buildReport(config)
	if err != nil {
		exitError(err)
	}
	if err := writeReport(config, report); err != nil {
		exitError(err)
	}
	if config.Check {
		if err := validateConfigReport(config, report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if report.Decision != "PASS" || report.Resolution != "EXACT" {
			os.Exit(1)
		}
	}
}

func validateConfigReport(config Config, report Report) error {
	var manifest Manifest
	if _, err := readJSON(config.ManifestPath, &manifest); err != nil {
		return err
	}
	var contract Contract
	if _, err := readJSON(config.ContractPath, &contract); err != nil {
		return err
	}
	program, err := os.ReadFile(config.ProgramPath)
	if err != nil {
		return err
	}
	return validateReport(report, manifest, contract, program)
}

func parseConfig() Config {
	var config Config
	var dependencies string
	flag.StringVar(&config.RunPath, "run", "", "source workflow run JSON")
	flag.StringVar(&config.JobsPath, "jobs", "", "source workflow jobs JSON")
	flag.StringVar(&config.ManifestPath, "manifest", "examples/ci-effort-observation/operations.json", "checked-in operation manifest")
	flag.StringVar(&config.ContractPath, "contract", "examples/ci-effort-observation/contract.json", "checked-in Gooo contract")
	flag.StringVar(&config.ProgramPath, "program", "examples/ci-effort-observation/main.gooo", "checked-in Gooo program")
	flag.StringVar(&config.SummaryPath, "ci-summary", "", "source CI summary artifact")
	flag.StringVar(&config.EvidencePath, "ci-evidence", "", "source CI evidence artifact")
	flag.StringVar(&config.RepositoryStatusPath, "repository-status", "", "before/after repository status receipt")
	flag.StringVar(&config.OpenTofuPath, "opentofu-report", "", "OpenTofu report from the exact head")
	flag.StringVar(&config.OpenTofuMetaPath, "opentofu-artifact", "", "OpenTofu Actions artifact metadata")
	flag.StringVar(&config.PriorPath, "prior", "", "optional immutable prior receipt")
	flag.StringVar(&dependencies, "dependency-files", "go.mod,go.sum", "comma-separated dependency graph files")
	flag.StringVar(&config.Environment, "environment", "", "canonical environment allowlist descriptor")
	flag.StringVar(&config.OutputPath, "output", "", "JSON report path")
	flag.StringVar(&config.MarkdownPath, "markdown", "", "human-readable report path")
	flag.BoolVar(&config.Check, "check", false, "require PASS/EXACT")
	flag.Parse()
	for _, path := range strings.Split(dependencies, ",") {
		if strings.TrimSpace(path) != "" {
			config.DependencyFiles = append(config.DependencyFiles, strings.TrimSpace(path))
		}
	}
	return config
}

func writeReport(config Config, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if config.OutputPath == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = writeOwned(config.OutputPath, data)
	}
	if err != nil {
		return err
	}
	if config.MarkdownPath != "" {
		if err := writeOwned(config.MarkdownPath, []byte(report.HumanReport)); err != nil {
			return err
		}
	}
	return nil
}

func writeOwned(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return nil
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
