package main

import (
	"image"
)

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
