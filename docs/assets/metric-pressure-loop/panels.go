package main

import "image"

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

func selectedPressure(index int, state phaseState) bool {
	if state.epoch == 1 {
		return (index == 0 || index == 1) && state.selection > 0.55
	}
	return (index == 1 || index == 3) && state.selection > 0.55
}

func drawAgents(img *image.Paletted, state phaseState) {
	box(img, 540, 150, 900, 500, panel, violet)
	drawText(img, 556, 170, "100 PARALLEL AGENTS", 2, textPrimary)
	drawText(img, 556, 197, "HEURISTIC  /  LOCAL INFERENCE ALLOWED", 1, violet)
	drawText(img, 556, 214, "10 LANES X 10  |  EXAMPLE WORKLOAD", 1, textMuted)
	for i := 0; i < 100; i++ {
		col, row := i%10, i/10
		x, y := 568+col*19, 226+row*19
		launched := state.attempts > float64(i+1)/108
		if !launched {
			circle(img, x+8, y+8, 6, dark)
			stroke(img, x+2, y+2, x+14, y+14, border)
			continue
		}
		switch i % 7 {
		case 0, 1:
			circle(img, x+8, y+8, 6, coral)
			cross(img, x+3, y+3, x+13, y+13, dark)
			cross(img, x+13, y+3, x+3, y+13, dark)
		case 2:
			diamond(img, x+8, y+8, 7, unknown)
		default:
			circle(img, x+8, y+8, 6, teal)
		}
		if i%13 == 0 && state.accepted > 0.15 {
			stroke(img, x+1, y+1, x+15, y+15, amber)
		}
	}
	drawCircleLegend(img, 568, 424, teal, "PASS")
	drawCrossLegend(img, 626, 424, coral, "FAIL")
	drawDiamondLegend(img, 684, 424, unknown, "UNKNOWN")
	drawPressureGauge(img, 568, 449, "PERFORMANCE", agentPerformance(state), amber)
	drawPressureGauge(img, 568, 468, "COMPLETENESS", agentCompleteness(state), cyan)
	drawText(img, 568, 488, "CIRCLE = AGENT  |  BOX = POLICY OR CI", 1, textMuted)
}

func agentPerformance(state phaseState) float64 {
	if state.reject > 0 {
		if state.reject < 0.5 {
			return 0.52 + 0.36*state.reject*2
		}
		return 0.88 - 0.36*(state.reject-0.5)*2
	}
	if state.accepted > 0 {
		return 0.52 + 0.38*state.accepted
	}
	return state.ceiling
}

func agentCompleteness(state phaseState) float64 {
	if state.reject > 0 {
		if state.reject < 0.5 {
			return 0.52 - 0.30*state.reject*2
		}
		return 0.22 + 0.30*(state.reject-0.5)*2
	}
	if state.accepted > 0 {
		return 0.52 + 0.36*state.accepted
	}
	return state.ceiling
}

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

func drawCrossLegend(img *image.Paletted, x, y int, colorIndex uint8, label string) {
	cross(img, x+2, y+2, x+12, y+12, colorIndex)
	cross(img, x+12, y+2, x+2, y+12, colorIndex)
	drawText(img, x+18, y+2, label, 1, textMuted)
}

func drawDiamondLegend(img *image.Paletted, x, y int, colorIndex uint8, label string) {
	diamond(img, x+7, y+7, 7, colorIndex)
	drawText(img, x+18, y+2, label, 1, textMuted)
}
