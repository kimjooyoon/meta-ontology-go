package main

import (
	"bytes"
	"image/gif"
	"image/png"
	"os"
	"testing"
)

func TestGeneratedAssetsAreCurrent(t *testing.T) {
	wantGIF, wantPNG, err := encodedAssets()
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		name string
		want []byte
	}{
		{name: "metric-pressure-loop.gif", want: wantGIF},
		{name: "metric-pressure-loop.png", want: wantPNG},
	} {
		got, err := os.ReadFile(fixture.name)
		if err != nil {
			t.Fatalf("read %s: %v; regenerate with go run ./docs/assets/metric-pressure-loop", fixture.name, err)
		}
		if !bytes.Equal(got, fixture.want) {
			t.Fatalf("%s is stale; regenerate with go run ./docs/assets/metric-pressure-loop", fixture.name)
		}
	}
}
func TestGeneratedAnimationMetadata(t *testing.T) {
	gifData, pngData, err := encodedAssets()
	if err != nil {
		t.Fatal(err)
	}
	animation, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		t.Fatalf("decode GIF: %v", err)
	}
	if got := len(animation.Image); got != frameCount {
		t.Fatalf("GIF frame count = %d, want %d", got, frameCount)
	}
	if animation.LoopCount != 0 {
		t.Fatalf("GIF loop count = %d, want 0 (forever)", animation.LoopCount)
	}
	for i, delay := range animation.Delay {
		if delay != frameDelay {
			t.Fatalf("GIF frame %d delay = %d, want %d hundredths", i, delay, frameDelay)
		}
	}
	if got := animation.Config.Width; got != width {
		t.Fatalf("GIF width = %d, want %d", got, width)
	}
	if got := animation.Config.Height; got != height {
		t.Fatalf("GIF height = %d, want %d", got, height)
	}
	if len(gifData) >= 5*1024*1024 {
		t.Fatalf("GIF size = %d bytes, want less than 5 MiB", len(gifData))
	}
	duration := len(animation.Delay) * frameDelay
	if duration < 22*100 || duration > 30*100 {
		t.Fatalf("GIF duration = %d hundredths, want 22 to 30 seconds", duration)
	}
	pngConfig, err := png.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("decode PNG config: %v", err)
	}
	if pngConfig.Width != width || pngConfig.Height != height {
		t.Fatalf("PNG dimensions = %dx%d, want %dx%d", pngConfig.Width, pngConfig.Height, width, height)
	}
}
