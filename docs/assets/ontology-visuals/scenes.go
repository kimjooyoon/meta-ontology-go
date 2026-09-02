package main

import "image"

type sceneRenderer func(*image.Paletted, manifestScene, int)

var renderers = map[string]sceneRenderer{
	"ontology.visual.intent-ir-lowering":        renderIntentIR,
	"ontology.visual.authority-boundary":        renderAuthorityBoundary,
	"ontology.visual.munchausen-proof-choice":   renderProofChoice,
	"ontology.visual.claim-evidence-lifecycle":  renderClaimLifecycle,
	"ontology.visual.unknown-cause-descent":     renderUnknownDescent,
	"ontology.visual.precedence-counterexample": renderPrecedence,
	"ontology.visual.package-resolution":        renderPackageResolution,
	"ontology.visual.incremental-conformance":   renderIncremental,
	"ontology.visual.bootstrap-oracle":          renderBootstrap,
	"ontology.visual.promotion-lineage":         renderPromotion,
}

func renderFrame(scene manifestScene, frame int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, 0, 0, width, height, background)
	drawText(img, 28, 20, "GOOO ONTOLOGY VISUALS", 2, textPrimary)
	drawText(img, 30, 47, scene.Title, 1, cyan)
	fill(img, 742, 18, 932, 50, panelRaised)
	stroke(img, 742, 18, 932, 50, violet)
	drawText(img, 756, 29, fmtSceneNumber(scene.Sequence), 1, violet)
	renderers[scene.ID](img, scene, frame)
	fill(img, 28, 480, 932, 512, panelRaised)
	drawText(img, 42, 488, "SOURCE-BACKED SEMANTICS", 1, textMuted)
	drawText(img, 250, 488, "LABELS + SHAPES CARRY STATE", 1, textMuted)
	drawText(img, 614, 488, "NO AGGREGATE SCORE", 1, amber)
	return img
}

func fmtSceneNumber(n int) string {
	if n < 10 {
		return "0" + itoa(n) + " / 10"
	}
	return itoa(n) + " / 10"
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(byte('0'+n%10)) + out
		n /= 10
	}
	return out
}
func caption(img *image.Paletted, value string) { drawText(img, 30, 445, value, 1, textMuted) }

func renderIntentIR(img *image.Paletted, _ manifestScene, frame int) {
	p1, p2, p3 := progress(frame, 2, 14), progress(frame, 14, 27), progress(frame, 27, 42)
	drawNode(img, 34, 140, 238, 230, ".GOOO INTENT", []string{"activity PayOrder", "stable ID + source span", "business meaning"}, cyan, p1 > .1)
	drawNode(img, 360, 140, 238, 230, "SEMANTIC IR", []string{"normalized nodes", "used / generated", "canonical identity"}, violet, p2 > .1)
	drawNode(img, 686, 108, 238, 132, "GENERATED GO", []string{"structural regions", "handwritten slots"}, teal, p3 > .1)
	drawNode(img, 686, 258, 238, 132, "EVIDENCE", []string{"source spans", "facts + digest"}, amber, p3 > .1)
	flow(img, 272, 250, 360, 250, cyan, p1 > .2)
	flow(img, 598, 250, 686, 174, violet, p2 > .2)
	flow(img, 598, 250, 686, 324, violet, p2 > .6)
	drawText(img, 42, 388, "PARSE", 1, cyan)
	drawText(img, 170, 388, "->", 1, textMuted)
	drawText(img, 240, 388, "NORMALIZE", 1, violet)
	drawText(img, 426, 388, "->", 1, textMuted)
	drawText(img, 500, 388, "PROJECT + RECORD", 1, teal)
	caption(img, "One intent source lowers once; Go structure and proof are derived views.")
}

func renderAuthorityBoundary(img *image.Paletted, _ manifestScene, frame int) {
	p := progress(frame, 4, 35)
	drawNode(img, 42, 130, 262, 260, "HANDWRITTEN", []string{"AUTHORITY", "irreducible logic", "implementation slots", "owns behavior"}, cyan, p > .15)
	drawNode(img, 656, 130, 262, 120, "GENERATED VIEW", []string{"structural Go", "regenerate from IR"}, teal, p > .35)
	drawNode(img, 656, 290, 262, 120, "EVIDENCE VIEW", []string{"append-only facts", "provenance + CI"}, amber, p > .55)
	line(img, 485, 108, 485, 420, violet)
	drawText(img, 429, 126, "AUTHORITY", 1, violet)
	drawText(img, 430, 142, "BOUNDARY", 1, violet)
	flow(img, 304, 210, 656, 190, cyan, p > .2)
	flow(img, 304, 294, 656, 350, cyan, p > .45)
	cross(img, 330, 365, 360, 395, coral)
	cross(img, 360, 365, 330, 395, coral)
	drawText(img, 374, 374, "NO WRITE-BACK TO INTENT", 1, coral)
	badge(img, 396, 224, "SOURCE OF BEHAVIOR", cyan, 'o')
	badge(img, 396, 276, "DERIVED OBSERVATION", amber, 'd')
	caption(img, "Handwritten Go owns irreducible behavior; generated and evidence views cannot become intent.")
}

func renderProofChoice(img *image.Paletted, _ manifestScene, frame int) {
	choice := (frame / 12) % 3
	drawNode(img, 42, 184, 206, 150, "CLAIM", []string{"prove this transition", "choice is explicit"}, cyan, frame > 3)
	choices := []struct {
		x    int
		name string
		sub  []string
		c    uint8
	}{{300, "FOUNDATION", []string{"fix denominator", "reject unknown"}, teal}, {530, "COHERENCE", []string{"bind receipts", "same facts"}, violet}, {760, "REGRESSION", []string{"replay digest", "compare exact"}, amber}}
	for i, v := range choices {
		active := choice == i && frame > 7
		drawNode(img, v.x, 145, 170, 228, v.name, v.sub, v.c, active)
		flow(img, 248, 258, v.x, 258, v.c, active)
	}
	drawText(img, 48, 368, "MUNCHAUSEN QUESTION", 1, amber)
	drawText(img, 48, 388, "WHAT JUSTIFIES THIS CLAIM?", 2, textPrimary)
	drawText(img, 300, 395, "FOUNDATION = DENOMINATOR", 1, teal)
	drawText(img, 530, 395, "COHERENCE = RECEIPT CHAIN", 1, violet)
	drawText(img, 760, 395, "REGRESSION = REPLAY", 1, amber)
	caption(img, "The proof choice is part of the claim; it never authorizes repository mutation.")
}

func renderClaimLifecycle(img *image.Paletted, _ manifestScene, frame int) {
	p1, p2, p3 := progress(frame, 2, 13), progress(frame, 13, 25), progress(frame, 25, 43)
	drawNode(img, 38, 180, 200, 150, "CLAIM", []string{"expected relation", "stable subject"}, cyan, p1 > .1)
	drawNode(img, 380, 180, 200, 150, "EVIDENCE", []string{"source-backed", "fresh + bound"}, violet, p2 > .1)
	flow(img, 238, 255, 380, 255, cyan, p1 > .3)
	flow(img, 580, 255, 690, 180, violet, p2 > .4)
	flow(img, 580, 255, 690, 255, violet, p2 > .7)
	flow(img, 580, 255, 690, 330, violet, p2 > .9)
	stateCard(img, 690, 132, 230, 70, "CLOSED", []string{"proof complete"}, green, p3 > .55, 's')
	stateCard(img, 690, 222, 230, 70, "UNKNOWN", []string{"evidence absent"}, unknown, p3 > .25 && p3 < .8, 'd')
	stateCard(img, 690, 312, 230, 70, "REFUTED", []string{"counterexample"}, coral, p3 > .8, 'x')
	drawText(img, 40, 382, "OPEN", 1, cyan)
	drawText(img, 118, 382, "->", 1, textMuted)
	drawText(img, 178, 382, "OBSERVED", 1, violet)
	drawText(img, 300, 382, "->", 1, textMuted)
	drawText(img, 360, 382, "ONE EXPLICIT OUTCOME", 1, amber)
	caption(img, "Evidence closes a claim, missing evidence lowers resolution, and contradiction preserves a refutation.")
}

func stateCard(img *image.Paletted, x, y, w, h int, title string, lines []string, c uint8, active bool, shape byte) {
	fill(img, x, y, x+w, y+h, panel)
	stroke(img, x, y, x+w, y+h, activeColor(1, 0, 1, c))
	if shape == 'd' {
		diamond(img, x+22, y+35, 13, c)
	} else if shape == 'x' {
		cross(img, x+11, y+24, x+33, y+46, c)
	} else {
		fill(img, x+11, y+24, x+33, y+46, c)
	}
	drawText(img, x+48, y+14, title, 2, c)
	textLines(img, x+48, y+42, lines, 1, textPrimary)
	if active {
		stroke(img, x+2, y+2, x+w-2, y+h-2, c)
	}
}

func renderUnknownDescent(img *image.Paletted, _ manifestScene, frame int) {
	fields := []struct {
		title, value string
		c            uint8
	}{{"STAGE", "semantic-conformance", cyan}, {"STEP", "extractor-observe", violet}, {"REASON", "EVIDENCE ABSENT", amber}, {"CLASS", "DIRECT_MISSING", unknown}, {"NEXT_OPERATION", "restore-decomposition-evidence", teal}, {"BLOCKED_BY", "[split-source-metrics]", coral}}
	drawText(img, 42, 110, "UNKNOWN = A CAUSAL FRONTIER, NOT A SCORE", 1, amber)
	diamond(img, 886, 116, 15, unknown)
	drawText(img, 848, 140, "UNKNOWN", 1, unknown)
	for i, v := range fields {
		y := 160 + i*43
		active := frame >= i*6+4
		fill(img, 42, y, 700, y+32, panel)
		stroke(img, 42, y, 700, y+32, activeColor(1, 0, 1, v.c))
		drawText(img, 58, y+10, v.title, 1, v.c)
		drawText(img, 220, y+10, v.value, 1, textPrimary)
		if i < len(fields)-1 {
			flow(img, 720, y+16, 780, y+59, v.c, active)
		}
	}
	drawNode(img, 780, 206, 140, 100, "CAUSE", []string{"descend", "until repair"}, coral, frame > 24)
	caption(img, "The reason is carried downward into class, next operation, and the exact blockers.")
}

func renderPrecedence(img *image.Paletted, _ manifestScene, frame int) {
	drawText(img, 42, 110, "DECISION PRECEDENCE", 2, amber)
	drawText(img, 42, 138, "REFUTED > UNKNOWN > CLOSED", 1, textPrimary)
	stateCard(img, 54, 182, 230, 70, "REFUTED", []string{"priority 130", "known contradiction"}, coral, frame > 9, 'x')
	stateCard(img, 54, 272, 230, 70, "UNKNOWN", []string{"priority 60", "missing evidence"}, unknown, frame > 18, 'd')
	stateCard(img, 54, 362, 230, 70, "CLOSED", []string{"priority 0", "only if no blocker"}, green, frame > 29, 's')
	flow(img, 286, 217, 380, 217, coral, frame > 10)
	flow(img, 286, 307, 380, 307, unknown, frame > 19)
	flow(img, 286, 397, 380, 397, green, frame > 30)
	drawNode(img, 406, 172, 226, 250, "SELECTOR", []string{"highest blocker", "fail closed", "no averaging"}, violet, frame > 15)
	flow(img, 632, 220, 710, 220, violet, frame > 24)
	drawNode(img, 710, 172, 210, 112, "CLAIM", []string{"REFUTED", "returned"}, coral, frame > 24)
	drawNode(img, 710, 318, 210, 104, "COUNTEREXAMPLE", []string{"append-only", "blocker preserved"}, amber, frame > 31)
	caption(img, "A counterexample stays visible; state classes are ordered, never collapsed into a percentage.")
}

func renderPackageResolution(img *image.Paletted, _ manifestScene, frame int) {
	p := progress(frame, 2, 40)
	drawNode(img, 42, 144, 196, 112, "activity.gooo", []string{"package billing", "activity PayOrder"}, cyan, p > .1)
	drawNode(img, 42, 300, 196, 112, "entities.gooo", []string{"package billing", "Order + Receipt"}, cyan, p > .2)
	drawNode(img, 318, 190, 210, 176, "CANONICAL ORDER", []string{"activity.gooo", "entities.gooo", "sort once"}, violet, p > .3)
	drawNode(img, 608, 190, 174, 176, "PACKAGE UNIT", []string{"parse each", "combine", "one namespace"}, teal, p > .55)
	drawNode(img, 826, 190, 100, 176, "ENTRY", []string{"PayOrder", "execute"}, amber, p > .75)
	flow(img, 238, 200, 318, 230, cyan, p > .15)
	flow(img, 238, 356, 318, 326, cyan, p > .25)
	flow(img, 528, 278, 608, 278, violet, p > .45)
	flow(img, 782, 278, 826, 278, teal, p > .7)
	drawText(img, 320, 400, "PACKAGE_SOURCE_FILES 2/2", 1, teal)
	drawText(img, 590, 400, "DETERMINISTIC RECEIPT", 1, amber)
	caption(img, "Immediate .gooo files sort by canonical relative filename before one package bind and entry execution.")
}

func renderIncremental(img *image.Paletted, _ manifestScene, frame int) {
	drawNode(img, 42, 180, 190, 156, "CHANGED SURFACE", []string{"source span", "input digest", "scope"}, cyan, frame > 3)
	ops := []struct {
		y     int
		title string
		lines []string
		c     uint8
		shape byte
	}{{120, "EXECUTE", []string{"fresh input", "produce receipt"}, teal, 's'}, {200, "REUSE", []string{"exact digest", "replay receipt"}, violet, 'o'}, {280, "UNKNOWN", []string{"missing oracle", "lower resolution"}, unknown, 'd'}, {360, "REFUTED", []string{"counterexample", "retain blocker"}, coral, 'x'}}
	for i, v := range ops {
		active := frame > i*8+7
		stateCard(img, 330, v.y, 280, 70, v.title, v.lines, v.c, active, v.shape)
		flow(img, 232, 258, 330, v.y+35, v.c, active)
	}
	drawNode(img, 700, 190, 220, 170, "CONFORMANCE", []string{"operation result", "evidence-bound", "not aggregate"}, amber, frame > 28)
	flow(img, 610, 235, 700, 250, amber, frame > 28)
	flow(img, 610, 315, 700, 300, coral, frame > 36)
	caption(img, "Incremental conformance chooses EXECUTE, REUSE, UNKNOWN, or REFUTED per bound operation.")
}

func renderBootstrap(img *image.Paletted, _ manifestScene, frame int) {
	p := progress(frame, 3, 39)
	drawNode(img, 40, 190, 160, 108, ".GOOO", []string{"bounded input"}, cyan, p > .1)
	drawNode(img, 292, 125, 230, 150, "BOUNDED KERNEL", []string{"semantic lowering", "candidate view", "not authority"}, violet, p > .25)
	drawNode(img, 292, 300, 230, 150, "BOOTSTRAP ORACLE", []string{"independent verifier", "fallback witness"}, teal, p > .35)
	drawNode(img, 632, 210, 170, 150, "PARITY", []string{"compare digests", "same facts"}, amber, p > .58)
	drawNode(img, 834, 210, 92, 150, "GATE", []string{"NOT", "PROMOTED"}, coral, p > .82)
	flow(img, 200, 244, 292, 195, cyan, p > .15)
	flow(img, 200, 244, 292, 354, cyan, p > .28)
	flow(img, 522, 195, 632, 252, violet, p > .48)
	flow(img, 522, 354, 632, 318, teal, p > .6)
	flow(img, 802, 285, 834, 285, amber, p > .8)
	drawText(img, 42, 410, "CANDIDATE PARITY EVIDENCE", 1, amber)
	drawText(img, 440, 410, "EXTERNAL ORACLE REMAINS INDEPENDENT", 1, teal)
	caption(img, "A bounded self-hosted candidate is compared with an independent bootstrap oracle; parity is evidence, not promotion.")
}

func renderPromotion(img *image.Paletted, _ manifestScene, frame int) {
	labels := []struct {
		x     int
		title string
		lines []string
		c     uint8
	}{{38, "FEATURE PR", []string{"intent change", "same repository"}, cyan}, {224, "EXACT-HEAD CI", []string{"head SHA bound", "all checks"}, violet}, {410, "DEV", []string{"adopted", "protected"}, teal}, {596, "POST-ADOPTION", []string{"observe", "fresh evidence"}, amber}, {782, "MAIN PROMOTION", []string{"dev -> main", "fast-forward"}, green}}
	for i, v := range labels {
		active := frame > i*8+3
		drawNode(img, v.x, 190, 150, 150, v.title, v.lines, v.c, active)
		if i < len(labels)-1 {
			flow(img, v.x+150, 265, labels[i+1].x, 265, v.c, active)
		}
	}
	drawText(img, 42, 382, "FEATURE PR", 1, cyan)
	drawText(img, 212, 382, "->", 1, textMuted)
	drawText(img, 246, 382, "EXACT-HEAD CI", 1, violet)
	drawText(img, 416, 382, "-> DEV", 1, teal)
	drawText(img, 536, 382, "-> OBSERVE", 1, amber)
	drawText(img, 682, 382, "->", 1, textMuted)
	drawText(img, 742, 382, "MAIN", 1, green)
	drawText(img, 42, 420, "NO SKIP: DEV REF MUST BE EXACT BEFORE PROMOTION", 1, coral)
	caption(img, "Feature adoption and post-adoption evidence precede the exact-head dev-to-main promotion route.")
}
