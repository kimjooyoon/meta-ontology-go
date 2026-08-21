package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func drawHeader(img *image.Paletted, state phaseState) {
	drawText(img, 40, 28, "SEMANTIC SELF IMPROVEMENT LOOP", 2, textPrimary)
	drawText(img, 43, 78, "SOURCE BACKED GAINS  ->  VERIFIED NEXT FLOOR", 1, textMuted)
	fill(img, 650, 38, 832, 78, panelRaised)
	stroke(img, 650, 38, 832, 78, violet)
	stageLabels := []string{"EPOCH 1 FLOOR", "SELECT FOCUS", "100 PARALLEL ATTEMPTS", "REJECT NON COMPENSATING", "CI REQUALIFY", "PROOF PATHS", "RATCHET NEXT FLOOR", "EPOCH 2 START", "NEXT WORK OPENS"}
	drawText(img, 665, 52, stageLabels[state.stage], 1, violet)
	fill(img, 864, 28, 1240, 88, panelRaised)
	stroke(img, 864, 28, 1240, 88, amber)
	drawText(img, 887, 40, "DETERMINISTIC CI -", 2, amber)
	drawText(img, 887, 68, "NOT INFERENCE", 2, textPrimary)
	if state.sealed > 0.5 {
		fill(img, 650, 82, 833, 96, green)
		drawText(img, 660, 84, "NEXT FLOOR SEALED", 1, dark)
	}
}
func drawFlowArrows(img *image.Paletted, state phaseState) {
	line(img, 250, 315, 270, 315, border)
	arrow(img, 270, 315, cyan)
	line(img, 510, 315, 540, 315, border)
	arrow(img, 540, 315, violet)
	line(img, 900, 315, 930, 315, border)
	arrow(img, 930, 315, teal)
	line(img, 1090, 500, 1090, 540, border)
	arrow(img, 1090, 540, amber)
	line(img, 1090, 674, 105, 674, border)
	arrowLeft(img, 105, 674, cyan)
	if state.sealed > 0 {
		line(img, 105, 674, 105, 625, cyan)
		arrow(img, 105, 625, cyan)
	}
}
func writePreviews(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create preview directory: %w", err)
	}
	for _, frame := range []int{0, 24, 45, 60, 68, 78, 100, 120, staticFrame} {
		path := filepath.Join(dir, fmt.Sprintf("frame-%03d.png", frame))
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create preview %s: %w", path, err)
		}
		if err := png.Encode(file, renderFrame(frame)); err != nil {
			_ = file.Close()
			return fmt.Errorf("encode preview %s: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close preview %s: %w", path, err)
		}
	}
	return nil
}
