package main

import (
	"image"
)

func drawFloor(img *image.Paletted, state phaseState) {
	box(img, 35, 150, 250, 500, panel, cyan)
	drawText(img, 54, 170, "EPOCH FLOOR", 2, cyan)
	drawText(img, 54, 194, "ALL BASE METRICS GUARDED", 1, textPrimary)
	drawText(img, 54, 211, "N=6 EXAMPLE POLICY  |  OBSERVED", 1, textMuted)
	metrics := []struct {
		label  string
		weight float64
	}{
		{"LATENCY", 0.86},
		{"CORRECTNESS", 0.74},
		{"MEMORY", 0.68},
		{"SECURITY", 0.79},
		{"PROVENANCE", 0.72},
		{"COST", 0.63},
	}
	for i, metric := range metrics {
		y := 240 + i*34
		value := 0.18 + metric.weight*0.78*state.floor/0.90
		fill(img, 53, y, 232, y+24, panelRaised)
		drawText(img, 61, y+7, metric.label, 1, textPrimary)
		fill(img, 146, y+7, 227, y+16, dark)
		fill(img, 146, y+7, 146+int(81*clamp(value)), y+16, cyan)
		line(img, 146+int(81*clamp(state.floor/0.90)), y+3, 146+int(81*clamp(state.floor/0.90)), y+20, textPrimary)
	}
	drawText(img, 54, 452, "ALL N BASE METRICS", 1, teal)
	drawText(img, 54, 468, "NON REGRESSION FLOOR", 1, textPrimary)
	drawText(img, 54, 486, "OBSERVE  /  NEVER TRADE", 1, textMuted)
}
func drawSelector(img *image.Paletted, state phaseState) {
	box(img, 270, 150, 510, 500, panel, violet)
	drawText(img, 286, 170, "PROTECTED SELECTOR", 2, violet)
	drawText(img, 286, 197, "SYSTEM POLICY + SPI", 1, textPrimary)
	drawText(img, 286, 214, "DEPENDENCY GRAPH + FRESHNESS", 1, textMuted)
	drawText(img, 286, 231, "SLACK + PRIORITY ORDER", 1, textMuted)
	drawText(img, 286, 248, "SYSTEM, NOT AN LLM", 1, amber)
	fill(img, 286, 269, 494, 316, panelRaised)
	if state.epoch == 1 {
		drawText(img, 300, 278, "FOCUS SUBSET  /  EXAMPLE", 1, amber)
	} else {
		drawText(img, 300, 278, "NEXT METRIC SET + SPI", 1, amber)
	}
	drawText(img, 300, 292, "K=2 OF M=4", 2, textPrimary)
	if state.epoch == 1 {
		drawText(img, 300, 307, "N=6 FLOOR GUARDS RETAINED", 1, textMuted)
	} else {
		drawText(img, 300, 307, "REGISTER PATH PROOF OBLIGATIONS", 1, textMuted)
	}
	pressures := []string{"PERFORMANCE", "COMPLETENESS", "SECURITY", "FRESHNESS"}
	for i, pressure := range pressures {
		y := 327 + i*30
		active := selectedPressure(i, state)
		fill(img, 286, y, 494, y+22, panelRaised)
		if active {
			fill(img, 286, y, 303, y+22, violet)
			drawText(img, 292, y+6, "K", 1, dark)
			drawText(img, 312, y+6, pressure, 1, textPrimary)
		} else {
			drawText(img, 294, y+6, "-", 1, textMuted)
			drawText(img, 312, y+6, pressure, 1, textMuted)
		}
	}
	drawText(img, 286, 456, "ACCEPTANCE >= 2", 1, teal)
	drawText(img, 286, 470, "INDEPENDENT DIMENSIONS", 1, teal)
	drawText(img, 286, 486, "ALL N GUARDS STAY ON", 1, textPrimary)
}
