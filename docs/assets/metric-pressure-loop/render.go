package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

const (
	stageFloor = iota
	stageSelect
	stageFanout
	stageReject
	stageAccept
	stageProof
	stageRatchet
	stageEpochTwo
	stageNextWork
)

type phaseState struct {
	stage     int
	epoch     int
	floor     float64
	ceiling   float64
	selection float64
	attempts  float64
	reject    float64
	accepted  float64
	proof     float64
	sealed    float64
	epochTwo  float64
	nextWork  float64
}

func renderFrame(frame int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, 0, 0, width, height, background)
	state := loopState(frame)

	drawHeader(img, state)
	drawFlowArrows(img, state)
	drawFloor(img, state)
	drawSelector(img, state)
	drawAgents(img, state)
	drawProvenance(img, state)
	drawCI(img, state)
	return img
}

func loopState(frame int) phaseState {
	p := float64(frame) / float64(frameCount-1)
	s := phaseState{stage: stageFloor, epoch: 1, floor: 0.28, ceiling: 0.28}
	switch {
	case p < 0.10:
		s.floor = lerp(0.28, 0.52, segment(p, 0.00, 0.10))
		s.ceiling = s.floor
	case p < 0.22:
		s.stage = stageSelect
		s.floor, s.ceiling = 0.52, 0.52
		s.selection = segment(p, 0.10, 0.22)
	case p < 0.35:
		s.stage = stageFanout
		s.floor, s.ceiling = 0.52, 0.52
		s.selection, s.attempts = 1, segment(p, 0.22, 0.35)
	case p < 0.49:
		s.stage = stageReject
		s.floor, s.ceiling = 0.52, 0.52
		s.selection, s.attempts = 1, 1
		s.reject = segment(p, 0.35, 0.49)
	case p < 0.68:
		s.stage = stageAccept
		s.floor, s.attempts, s.accepted = 0.52, 1, segment(p, 0.49, 0.68)
		s.ceiling = lerp(0.52, 0.90, s.accepted)
	case p < 0.77:
		s.stage = stageProof
		s.floor, s.ceiling, s.attempts = 0.52, 0.90, 1
		s.proof = segment(p, 0.68, 0.77)
	case p < 0.86:
		s.stage = stageRatchet
		s.floor, s.ceiling, s.attempts, s.proof = 0.52, 0.90, 1, 1
		s.sealed = segment(p, 0.77, 0.86)
	case p < 0.94:
		s.stage, s.epoch = stageEpochTwo, 2
		s.floor, s.ceiling, s.attempts, s.proof, s.sealed = 0.90, 0.90, 1, 1, 1
		s.epochTwo = segment(p, 0.86, 0.94)
		s.selection = segment(p, 0.90, 0.94)
	default:
		s.stage, s.epoch = stageNextWork, 2
		s.floor, s.ceiling, s.attempts, s.proof, s.sealed = 0.90, 0.90, 1, 1, 1
		s.epochTwo, s.selection = 1, 1
		s.nextWork = segment(p, 0.94, 1.00)
	}
	return s
}

func segment(p, start, end float64) float64 {
	if p <= start {
		return 0
	}
	if p >= end {
		return 1
	}
	return smooth((p - start) / (end - start))
}

func smooth(v float64) float64 {
	return v * v * (3 - 2*v)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*clamp(t)
}

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
