package main

import (
	"image"
)

func drawCI(img *image.Paletted, state phaseState) {
	box(img, 35, 525, 1245, 680, panelRaised, amber)
	drawText(img, 54, 544, "CI REQUALIFICATION", 2, amber)
	drawText(img, 54, 569, "DECLARED CEILING VECTOR + EXACT EVIDENCE", 1, textPrimary)
	drawText(img, 54, 586, "ALL BASE FLOORS GUARDED  /  NO COMPENSATION", 1, textMuted)
	drawText(img, 54, 603, "MISSING ORACLE OR EVIDENCE = UNKNOWN", 1, textMuted)
	drawText(img, 400, 540, "PERF", 1, textMuted)
	drawText(img, 400, 560, "COMP", 1, textMuted)
	drawText(img, 400, 580, "ALL N FLOORS", 1, textMuted)
	drawCeilingBar(img, 451, 540, agentPerformance(state), amber)
	drawCeilingBar(img, 451, 560, agentCompleteness(state), cyan)
	drawCeilingBar(img, 451, 580, state.floor, teal)
	fill(img, 745, 540, 1215, 605, dark)
	switch {
	case state.reject > 0:
		stroke(img, 745, 540, 1215, 605, coral)
		cross(img, 765, 557, 785, 579, coral)
		cross(img, 785, 557, 765, 579, coral)
		drawText(img, 805, 551, "REJECTED", 2, coral)
		drawText(img, 805, 574, "NON COMPENSATING VECTOR", 1, textPrimary)
		drawText(img, 805, 590, "SINGLE METRIC GREEN LIGHT", 1, textMuted)
	case state.sealed > 0:
		stroke(img, 745, 540, 1215, 605, green)
		fill(img, 765, 557, 787, 579, green)
		drawText(img, 771, 560, "+", 2, dark)
		drawText(img, 805, 551, "QUALIFIED", 2, green)
		drawText(img, 805, 574, "ALL DIMENSIONS PASS TOGETHER", 1, textPrimary)
		drawText(img, 805, 590, "CEILING BECOMES NEXT FLOOR", 1, textMuted)
	case state.proof > 0:
		stroke(img, 745, 540, 1215, 605, amber)
		drawText(img, 765, 552, "REQUALIFYING FULL VECTOR", 1, amber)
		drawText(img, 765, 574, "PROOF AND EVIDENCE MUST CLOSE", 1, textPrimary)
		drawText(img, 765, 590, "NO PROMOTION YET", 1, textMuted)
	default:
		stroke(img, 745, 540, 1215, 605, border)
		drawText(img, 765, 552, "WAITING FOR FULL VECTOR", 1, amber)
		drawText(img, 765, 574, "AGENTS NEVER AUTHORIZE", 1, textPrimary)
		drawText(img, 765, 590, "THE CEILING", 1, textMuted)
	}
	drawText(img, 54, 638, "CEILING  ->  IMMUTABLE FLOOR", 1, textMuted)
	line(img, 222, 646, 422, 646, border)
	fill(img, 222, 642, 222+int(200*state.sealed), 650, green)
	drawText(img, 451, 638, "RATCHET: EPOCH 1 -> EPOCH 2", 1, cyan)
	drawText(img, 700, 641, stageMessage(state), 1, textPrimary)
}
func stageMessage(state phaseState) string {
	return []string{
		"OBSERVE; DO NOT TRADE THE FLOOR",
		"SELECTOR RETAINS ALL FLOOR GUARDS",
		"100 ATTEMPTS LEAVE PROVENANCE",
		"SINGLE METRIC GAINS ARE REJECTED",
		"MULTI PRESSURE GAIN -> REQUALIFY",
		"PROOF CLOSES REQUIRED PATHS",
		"VERIFIED CEILING -> NEXT FLOOR",
		"NEW SPI + NEXT PATH OBLIGATIONS",
		"AGENTS PROPOSE | EVIDENCE COMPOUNDS | CI DECIDES",
	}[state.stage]
}
func drawCeilingBar(img *image.Paletted, x, y int, value float64, colorIndex uint8) {
	fill(img, x, y, x+245, y+11, dark)
	fill(img, x, y, x+int(245*clamp(value)), y+11, colorIndex)
	line(img, x+int(245*0.8), y-3, x+int(245*0.8), y+15, textPrimary)
}
func drawCircleLegend(img *image.Paletted, x, y int, colorIndex uint8, label string) {
	circle(img, x+7, y+7, 6, colorIndex)
	drawText(img, x+18, y+2, label, 1, textMuted)
}
