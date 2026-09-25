package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	width       = 1280
	height      = 720
	frameCount  = 144
	frameDelay  = 16
	defaultDir  = "docs/assets/metric-pressure-loop"
	staticFrame = frameCount - 1
)

func main() {
	outputDir := flag.String("out-dir", defaultDir, "directory for the generated GIF and PNG")
	check := flag.Bool("check", false, "verify generated assets are current")
	previewDir := flag.String("preview-dir", "", "optional directory for representative PNG frames")
	flag.Parse()

	if *check {
		if err := checkAssets(*outputDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := writeAssets(*outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *previewDir != "" {
		if err := writePreviews(*previewDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
func writeAssets(dir string) error {
	gifData, pngData, err := encodedAssets()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metric-pressure-loop.gif"), gifData, 0o644); err != nil {
		return fmt.Errorf("write GIF: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metric-pressure-loop.png"), pngData, 0o644); err != nil {
		return fmt.Errorf("write PNG: %w", err)
	}
	return nil
}
