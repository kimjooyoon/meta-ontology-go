package main

import (
	"fmt"
	"io"
	"path/filepath"

	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.outputDir == "" || cfg.expectedHead == "" ||
		cfg.platformID == "" || cfg.runner == "" ||
		cfg.expectedGOOS == "" || cfg.expectedGOARCH == "" {
		return fmt.Errorf("all platform witness arguments are required")
	}
	format := "tar.gz"
	if cfg.expectedGOOS == "windows" {
		format = "zip"
	}
	target := release.Target{
		ID: cfg.platformID, Runner: cfg.runner,
		GOOS: cfg.expectedGOOS, GOARCH: cfg.expectedGOARCH,
		ArchiveFormat: format,
	}
	receipt, err := release.BuildPlatform(release.BuildInput{
		Root: cfg.root, OutputDir: cfg.outputDir,
		ExpectedHead: cfg.expectedHead, Target: target,
	})
	if err != nil {
		return err
	}
	path := filepath.Join(cfg.outputDir, target.ID+".receipt.json")
	if err := release.WritePlatformReceipt(path, receipt); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "toolchain-release-platform: %s PASS/EXACT\n", target.ID)
	return err
}
