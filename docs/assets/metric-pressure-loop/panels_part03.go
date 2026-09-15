package main

import (
	"image"
)

func drawPressureGauge(img *image.Paletted, x, y int, label string, value float64, colorIndex uint8) {
	drawText(img, x, y, label, 1, textMuted)
	fill(img, x+98, y+2, x+300, y+11, dark)
	fill(img, x+98, y+2, x+98+int(202*clamp(value)), y+11, colorIndex)
	line(img, x+98+145, y-3, x+98+145, y+16, textPrimary)
}
func drawProvenance(img *image.Paletted, state phaseState) {
	box(img, 930, 150, 1245, 500, panel, teal)
	drawText(img, 947, 170, "PROVENANCE + PATH PROOF", 2, teal)
	drawText(img, 947, 197, "APPEND ONLY  /  FAIL CLOSED", 1, textPrimary)
	drawText(img, 947, 214, "SOURCE BACKED EVIDENCE", 1, textMuted)
	drawOutcomeCard(img, 947, 244, "REJECT  PERF UP / COMP DOWN", state.reject > 0.10 && state.reject < 0.62)
	drawOutcomeCard(img, 947, 273, "REJECT  COMP UP / PERF DOWN", state.reject >= 0.62)
	rows := []string{"A-044  PASS", "A-045  FAIL", "A-046  UNKNOWN"}
	colors := []uint8{teal, coral, unknown}
	for i, row := range rows {
		y := 316 + i*25
		visible := state.attempts > float64(i+2)/5
		fill(img, 948, y, 1226, y+18, dark)
		if visible {
			fill(img, 948, y, 959, y+18, colors[i])
			drawText(img, 968, y+5, row, 1, textPrimary)
		}
	}
	drawText(img, 947, 398, "REQUIRED PATH SET", 1, amber)
	for i, label := range []string{"P1", "P2", "P3"} {
		x := 948 + i*88
		fill(img, x, 413, x+72, 434, dark)
		progress := state.proof * float64(i+1) / 3
		if progress > 0 {
			fill(img, x, 413, x+int(72*clamp(progress)), 434, green)
		}
		stroke(img, x, 413, x+72, 434, border)
		drawText(img, x+8, 420, label, 1, textPrimary)
		if progress > 0.95 {
			drawText(img, x+35, 420, "+", 1, dark)
		}
	}
	drawText(img, 947, 454, "PROOF COMPUTES VIABLE PATHS", 1, amber)
	drawText(img, 947, 470, "USERS DO NOT INFER COMPLETION", 1, textPrimary)
	drawText(img, 947, 486, "MISSING ORACLE = UNKNOWN", 1, textMuted)
}
func drawOutcomeCard(img *image.Paletted, x, y int, label string, active bool) {
	fill(img, x, y, 1226, y+22, panelRaised)
	if active {
		stroke(img, x, y, 1226, y+22, coral)
		cross(img, x+8, y+5, x+16, y+13, coral)
		cross(img, x+16, y+5, x+8, y+13, coral)
		drawText(img, x+27, y+6, label, 1, textPrimary)
		return
	}
	drawText(img, x+9, y+6, label, 1, textMuted)
}
