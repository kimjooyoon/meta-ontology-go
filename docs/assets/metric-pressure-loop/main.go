package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
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

func checkAssets(dir string) error {
	wantGIF, wantPNG, err := encodedAssets()
	if err != nil {
		return err
	}
	for _, file := range []struct {
		name string
		want []byte
	}{
		{name: "metric-pressure-loop.gif", want: wantGIF},
		{name: "metric-pressure-loop.png", want: wantPNG},
	} {
		path := filepath.Join(dir, file.name)
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w; run go run ./docs/assets/metric-pressure-loop", path, readErr)
		}
		if !bytes.Equal(got, file.want) {
			return fmt.Errorf("generated asset is stale: %s; run go run ./docs/assets/metric-pressure-loop", path)
		}
	}
	return nil
}

func encodedAssets() ([]byte, []byte, error) {
	animation := &gif.GIF{
		Image:     make([]*image.Paletted, 0, frameCount),
		Delay:     make([]int, 0, frameCount),
		LoopCount: 0,
	}
	for frame := 0; frame < frameCount; frame++ {
		animation.Image = append(animation.Image, renderFrame(frame))
		animation.Delay = append(animation.Delay, frameDelay)
	}
	var gifBuffer bytes.Buffer
	if err := gif.EncodeAll(&gifBuffer, animation); err != nil {
		return nil, nil, fmt.Errorf("encode GIF: %w", err)
	}
	var pngBuffer bytes.Buffer
	if err := png.Encode(&pngBuffer, renderFrame(staticFrame)); err != nil {
		return nil, nil, fmt.Errorf("encode PNG: %w", err)
	}
	return gifBuffer.Bytes(), pngBuffer.Bytes(), nil
}
