package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type phaseState struct {
	selection float64
	attempts  float64
	ci        float64
	sealed    float64
}

func renderFrame(frame int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, 0, 0, width, height, background)
	state := phase(frame)

	drawHeader(img, state)
	drawFlowArrows(img, state)
	drawFloor(img, state)
	drawSelector(img, state)
	drawAgents(img, state)
	drawProvenance(img, state)
	drawCI(img, state)
	return img
}

func phase(frame int) phaseState {
	p := float64(frame%frameCount) / float64(frameCount)
	s := phaseState{}
	s.selection = segment(p, 0.12, 0.25)
	s.attempts = segment(p, 0.25, 0.68)
	s.ci = segment(p, 0.68, 0.84)
	s.sealed = segment(p, 0.84, 0.94)
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

func drawHeader(img *image.Paletted, state phaseState) {
	drawText(img, 40, 28, "SEMANTIC PRESSURE LOOP", 3, textPrimary)
	drawText(img, 43, 78, "POLICY-DEFINED CEILINGS  |  APPEND-ONLY EVIDENCE", 1, textMuted)
	fill(img, 864, 28, 1240, 88, panelRaised)
	stroke(img, 864, 28, 1240, 88, amber)
	drawText(img, 887, 40, "DETERMINISTIC CI -", 2, amber)
	drawText(img, 887, 68, "NOT INFERENCE", 2, textPrimary)
	if state.sealed > 0.5 {
		fill(img, 735, 82, 833, 96, green)
		drawText(img, 744, 84, "NEXT FLOOR SEALED", 1, dark)
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

func drawFloor(img *image.Paletted, state phaseState) {
	box(img, 35, 150, 250, 500, panel, cyan)
	drawText(img, 54, 170, "FLOOR", 2, cyan)
	drawText(img, 54, 194, "STABLE BASE METRICS", 1, textPrimary)
	drawText(img, 54, 211, "N=6  |  OBSERVED", 1, textMuted)
	metrics := []string{"LATENCY", "CORRECTNESS", "MEMORY", "SECURITY", "PROVENANCE", "COST"}
	for i, metric := range metrics {
		y := 240 + i*36
		fill(img, 53, y, 232, y+25, panelRaised)
		fill(img, 53, y, 58, y+25, cyan)
		drawText(img, 67, y+7, metric, 1, textPrimary)
		drawText(img, 190, y+7, "SEEN", 1, teal)
	}
	pulse := int(4 * (0.5 + 0.5*float64Sin(state.sealed*3.14159)))
	fill(img, 54, 470-pulse, 232, 474+pulse, cyan)
	drawText(img, 54, 479, "FLOOR CANNOT BE TRADED AWAY", 1, textPrimary)
}

func drawSelector(img *image.Paletted, state phaseState) {
	box(img, 270, 150, 510, 500, panel, violet)
	drawText(img, 286, 170, "PROTECTED SELECTOR", 2, violet)
	drawText(img, 286, 197, "SYSTEM POLICY + SPI", 1, textPrimary)
	drawText(img, 286, 214, "DEPENDENCY GRAPH", 1, textMuted)
	drawText(img, 286, 231, "EVIDENCE FRESHNESS / SLACK", 1, textMuted)
	drawText(img, 286, 248, "PRIORITY", 1, textMuted)
	fill(img, 286, 269, 494, 303, panelRaised)
	drawText(img, 300, 278, "THIS SYSTEM POLICY", 1, amber)
	drawText(img, 300, 292, "N=6  M=4  K=2", 2, textPrimary)
	pressures := []string{"PERFORMANCE", "COMPLETENESS", "SECURITY", "FRESHNESS"}
	for i, pressure := range pressures {
		y := 322 + i*34
		active := i < 2
		if active && state.selection < float64(i+1)/2 {
			active = false
		}
		fill(img, 286, y, 494, y+25, panelRaised)
		if active {
			fill(img, 286, y, 303, y+25, violet)
			drawText(img, 312, y+7, "K", 1, dark)
			drawText(img, 327, y+7, pressure, 1, textPrimary)
		} else {
			drawText(img, 296, y+7, pressure, 1, textMuted)
		}
	}
	drawText(img, 286, 456, "ACCEPTANCE >= 2", 1, teal)
	drawText(img, 286, 470, "INDEPENDENT DIMENSIONS", 1, teal)
	drawText(img, 286, 486, "SYSTEM, NOT AN LLM", 1, textPrimary)
}

func drawAgents(img *image.Paletted, state phaseState) {
	box(img, 540, 150, 900, 500, panel, violet)
	drawText(img, 556, 170, "100 PARALLEL AGENTS", 2, textPrimary)
	drawText(img, 556, 197, "HEURISTIC ATTEMPTS", 1, violet)
	const tile, gap = 18, 3
	for i := 0; i < 100; i++ {
		col, row := i%10, i/10
		x, y := 568+col*(tile+gap), 220+row*(tile+gap)
		progress := float64(i%17) / 17
		stateColor := unknown
		if state.attempts > progress {
			switch i % 7 {
			case 0, 1:
				stateColor = coral
			case 2:
				stateColor = unknown
			default:
				stateColor = teal
			}
		}
		fill(img, x, y, x+tile, y+tile, stateColor)
		if state.attempts > progress && i%13 == 0 {
			stroke(img, x, y, x+tile, y+tile, amber)
		}
	}
	drawText(img, 568, 470, "PASS", 1, teal)
	fill(img, 602, 472, 616, 486, teal)
	drawText(img, 630, 470, "FAIL", 1, coral)
	fill(img, 660, 472, 674, 486, coral)
	drawText(img, 688, 470, "UNKNOWN", 1, textMuted)
	fill(img, 750, 472, 764, 486, unknown)
	drawText(img, 568, 488, "INDIVIDUAL ATTEMPTS NEVER AUTHORIZE THE CEILING", 1, textMuted)
	drawPressureGauge(img, 568, 430, "PERFORMANCE", state.attempts, amber)
	drawPressureGauge(img, 568, 451, "COMPLETENESS", 1-state.attempts*0.68, cyan)
}

func drawPressureGauge(img *image.Paletted, x, y int, label string, value float64, colorIndex uint8) {
	drawText(img, x, y, label, 1, textMuted)
	fill(img, x+98, y+2, x+300, y+11, dark)
	fill(img, x+98, y+2, x+98+int(202*clamp(value)), y+11, colorIndex)
	line(img, x+98+145, y-3, x+98+145, y+16, textPrimary)
}

func drawProvenance(img *image.Paletted, state phaseState) {
	box(img, 930, 150, 1245, 500, panel, teal)
	drawText(img, 947, 170, "PROVENANCE", 2, teal)
	drawText(img, 947, 197, "APPEND-ONLY RECORDS", 1, textPrimary)
	drawText(img, 947, 214, "NO OVERWRITE", 1, textMuted)
	rows := []string{"A-042  UNKNOWN", "A-043  FAIL", "A-044  PASS", "A-045  PASS", "A-046  FAIL", "A-047  PASS", "A-048  UNKNOWN"}
	for i, row := range rows {
		y := 252 + i*31
		visible := state.attempts > float64(i)/float64(len(rows))*0.85
		if !visible {
			fill(img, 948, y, 1226, y+21, dark)
			continue
		}
		fill(img, 948, y, 1226, y+21, panelRaised)
		fill(img, 948, y, 959, y+21, []uint8{unknown, coral, teal, teal, coral, teal, unknown}[i])
		drawText(img, 968, y+6, row, 1, textPrimary)
	}
	drawText(img, 947, 468, "EVIDENCE FOLLOWS", 1, amber)
	drawText(img, 947, 481, "EVERY ATTEMPT", 1, amber)
}

func drawCI(img *image.Paletted, state phaseState) {
	box(img, 35, 525, 1245, 680, panelRaised, amber)
	drawText(img, 54, 544, "CI REQUALIFICATION", 2, amber)
	drawText(img, 54, 569, "FULL DECLARED CEILING VECTOR", 1, textPrimary)
	drawText(img, 54, 586, "ALL REQUIRED CEILING DIMENSIONS PASS TOGETHER", 1, textMuted)
	drawText(img, 400, 544, "PERF", 1, textMuted)
	drawText(img, 400, 564, "FEATURE", 1, textMuted)
	drawCeilingBar(img, 451, 544, state.ci, amber)
	drawCeilingBar(img, 451, 564, state.ci*0.9+0.1, cyan)
	fill(img, 745, 548, 1215, 603, dark)
	if state.ci > 0.82 {
		stroke(img, 745, 548, 1215, 603, green)
		fill(img, 764, 562, 786, 584, green)
		drawText(img, 770, 565, "+", 2, dark)
		drawText(img, 802, 561, "ALL REQUIRED DIMENSIONS", 2, green)
		drawText(img, 802, 581, "PASS TOGETHER", 2, textPrimary)
	} else {
		stroke(img, 745, 548, 1215, 603, border)
		drawText(img, 764, 562, "WAITING FOR FULL VECTOR REQUALIFICATION", 1, amber)
		drawText(img, 764, 580, "NO SINGLE-METRIC GREEN LIGHT", 1, textMuted)
	}
	drawText(img, 54, 638, "SEALED CEILING", 1, textMuted)
	line(img, 157, 646, 422, 646, border)
	fill(img, 157, 642, 157+int(265*state.sealed), 650, green)
	drawText(img, 451, 638, "-> NEXT FLOOR", 2, cyan)
	drawText(img, 672, 641, "SELECTOR OPENS THE NEXT PARALLEL SET", 1, textPrimary)
}

func drawCeilingBar(img *image.Paletted, x, y int, value float64, colorIndex uint8) {
	fill(img, x, y, x+245, y+11, dark)
	fill(img, x, y, x+int(245*clamp(value)), y+11, colorIndex)
	line(img, x+int(245*0.8), y-3, x+int(245*0.8), y+15, textPrimary)
}

func writePreviews(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create preview directory: %w", err)
	}
	for _, frame := range []int{0, frameCount / 2, staticFrame} {
		path := filepath.Join(dir, fmt.Sprintf("frame-%02d.png", frame))
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
