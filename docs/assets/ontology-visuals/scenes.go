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
	"ontology.visual.bootstrap-oracle":           renderBootstrap,
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
	merge := progress(frame, 2, 15)
	lower := progress(frame, 12, 25)
	materialize := progress(frame, 23, 35)
	drawNode(img, 32, 154, 214, 142, ".GOOO INTENT", []string{"PayOrder", "intent tokens", "source span"}, cyan, frame > 2)
	drawNode(img, 356, 154, 214, 142, "SEMANTIC IR", []string{"normalized", "canonical ID", "one meaning"}, violet, frame > 13)
	drawNode(img, 690, 112, 224, 112, "GENERATED GO", []string{"structural view", "handwritten slots"}, teal, frame > 27)
	drawNode(img, 690, 270, 224, 112, "EVIDENCE", []string{"facts + digest", "source-bound"}, amber, frame > 30)
	flow(img, 246, 225, 356, 225, cyan, frame > 5)
	flow(img, 570, 225, 690, 168, violet, frame > 23)
	flow(img, 570, 225, 690, 326, violet, frame > 27)
	drawText(img, 266, 203, "LOWER", 1, cyan)
	drawText(img, 600, 190, "PROJECT", 1, teal)
	drawText(img, 600, 288, "RECORD", 1, amber)
	for i := range 3 {
		x := 72 + i*48
		y := 238 - i*18
		if merge > float64(i)/3 {
			circle(img, moveInt(x, 448, merge), y, 9, cyan)
		}
	}
	if lower > 0.05 {
		circle(img, moveInt(448, 518, lower), 225, 9, violet)
	}
	if materialize > 0.05 {
		circle(img, moveInt(638, 754, materialize), moveInt(225, 168, materialize), 8, teal)
		circle(img, moveInt(638, 754, materialize), moveInt(225, 326, materialize), 8, amber)
	}
	drawText(img, 52, 346, "TOKENS MERGE", 1, cyan)
	drawText(img, 52, 370, "INTENT", 2, cyan)
	drawText(img, 210, 370, "->", 2, textMuted)
	drawText(img, 252, 370, "IR", 2, violet)
	drawText(img, 320, 370, "->", 2, textMuted)
	drawText(img, 362, 370, "VIEWS", 2, teal)
	drawText(img, 52, 402, "ONE SOURCE / TWO DERIVED RECEIPTS", 1, textPrimary)
	caption(img, "Intent tokens merge into semantic IR; generated structure and evidence materialize afterward.")
}

func renderAuthorityBoundary(img *image.Paletted, _ manifestScene, frame int) {
	forward := progress(frame, 2, 18)
	reverse := progress(frame, 15, 31)
	drawNode(img, 34, 160, 236, 184, "HANDWRITTEN", []string{"AUTHORITY", "behavior", "implementation", "slots"}, cyan, frame > 2)
	drawNode(img, 690, 118, 220, 110, "GENERATED GO", []string{"derived view", "regenerate"}, teal, frame > 10)
	drawNode(img, 690, 282, 220, 110, "EVIDENCE", []string{"append only", "observed facts"}, amber, frame > 14)
	fill(img, 454, 112, 508, 408, panelRaised)
	stroke(img, 454, 112, 508, 408, violet)
	drawText(img, 463, 176, "AUTHORITY", 1, violet)
	drawText(img, 463, 192, "WALL", 2, violet)
	flow(img, 270, 214, 690, 172, cyan, forward > 0.1)
	flow(img, 270, 286, 690, 336, cyan, forward > 0.35)
	if reverse > 0.05 {
		line(img, 690, 172, 508, 172, coral)
		arrowLeft(img, moveInt(690, 508, reverse), 172, coral)
	}
	if reverse > 0.75 {
		cross(img, 478, 152, 514, 188, coral)
		cross(img, 514, 152, 478, 188, coral)
		drawText(img, 520, 154, "BLOCKED", 1, coral)
	}
	if frame > 27 {
		circle(img, 481, 258, 12, cyan)
		drawText(img, 522, 252, "SOURCE PRESERVED", 1, cyan)
	}
	drawText(img, 62, 382, "FORWARD", 1, teal)
	drawText(img, 182, 382, "DERIVES STRUCTURE + FACTS", 1, textPrimary)
	drawText(img, 62, 408, "REVERSE WRITE", 1, coral)
	drawText(img, 182, 408, "HITS WALL / NO MUTATION", 1, textPrimary)
	caption(img, "Derived views can flow outward; a reverse-write attempt stops at the authority wall.")
}

func renderProofChoice(img *image.Paletted, _ manifestScene, frame int) {
	choice := 0
	claimP := progress(frame, 1, 9)
	chooseP := progress(frame, 7, 19)
	drawNode(img, 38, 190, 182, 116, "CLAIM", []string{"open", "needs proof"}, cyan, claimP > 0.1)
	drawText(img, 266, 112, "1 OF 3", 2, amber)
	diamond(img, 316, 248, 34, amber)
	drawText(img, 301, 238, "ONE", 1, dark)
	choices := []struct {
		y    int
		name string
		sub  []string
		c    uint8
	}{{130, "FOUNDATION", []string{"denominator", "reject unknown"}, teal}, {238, "COHERENCE", []string{"receipts", "same facts"}, violet}, {346, "REGRESSION", []string{"replay", "exact digest"}, amber}}
	flow(img, 220, 248, 282, 248, cyan, claimP > 0.4)
	if chooseP > 0.05 {
		circle(img, moveInt(282, 316, chooseP), 248, 8, amber)
	}
	for i, v := range choices {
		active := frame > 13 && choice == i
		stateCard(img, 430, v.y, 470, 76, v.name, v.sub, v.c, active, 's')
		if active {
			flow(img, 350, 248, 430, v.y+38, v.c, true)
		}
	}
	drawText(img, 42, 388, "MUNCHAUSEN", 1, amber)
	drawText(img, 42, 410, "CLAIM -> ONE EXPLICIT PROOF CHOICE", 1, textPrimary)
	caption(img, "Foundation, coherence, and regression are alternatives; exactly one proof obligation binds the claim.")
}

func renderClaimLifecycle(img *image.Paletted, _ manifestScene, frame int) {
	terminal := (frame / 12) % 3
	evidenceP := progress(frame, 1, 15)
	routeP := progress(frame, 12, 24)
	drawNode(img, 32, 190, 186, 118, "CLAIM", []string{"expected", "relation"}, cyan, frame > 2)
	drawNode(img, 280, 190, 186, 118, "EVIDENCE", []string{"bound", "fresh facts"}, violet, frame > 10)
	box(img, 532, 194, 656, 310, panelRaised, amber)
	drawText(img, 548, 210, "DECISION", 2, amber)
	drawText(img, 560, 246, "SWITCH", 2, amber)
	drawText(img, 554, 282, "ONE EXIT", 1, textPrimary)
	flow(img, 218, 249, 280, 249, cyan, evidenceP > 0.15)
	flow(img, 466, 249, 532, 249, violet, routeP > 0.1)
	if routeP > 0.1 {
		circle(img, moveInt(466, 590, routeP), 249, 9, violet)
	}
	states := []struct {
		y     int
		title string
		lines []string
		c     uint8
		shape byte
	}{{128, "CLOSED", []string{"proof complete"}, green, 's'}, {228, "UNKNOWN", []string{"evidence absent"}, unknown, 'd'}, {328, "REFUTED", []string{"counterexample"}, coral, 'x'}}
	for i, v := range states {
		active := frame > 21 && terminal == i
		stateCard(img, 700, v.y, 220, 78, v.title, v.lines, v.c, active, v.shape)
		if active {
			flow(img, 656, 249, 700, v.y+39, v.c, true)
		}
	}
	drawText(img, 46, 366, "EVIDENCE TOKEN", 1, violet)
	drawText(img, 46, 390, "-> DECISION SWITCH -> ONE TERMINAL SLOT", 1, textPrimary)
	caption(img, "The evidence token enters an exclusive switch; only one terminal relation is active.")
}

func renderUnknownDescent(img *image.Paletted, _ manifestScene, frame int) {
	frontier := []struct {
		x, y, w int
		title, value string
		c            uint8
	}{{56, 138, 148, "STAGE", "conformance", cyan}, {202, 178, 148, "STEP", "observe", violet}, {348, 218, 148, "REASON", "no evidence", amber}, {494, 258, 148, "CLASS", "direct miss", unknown}, {640, 298, 178, "NEXT", "restore", teal}, {788, 338, 132, "BLOCKED", "source", coral}}
	drawText(img, 40, 104, "UNKNOWN", 2, unknown)
	drawText(img, 190, 104, "CAUSAL FRONTIER", 2, amber)
	for i, node := range frontier {
		active := frame >= 3+i*5
		smallNode(img, node.x, node.y, node.w, 54, node.title, node.value, node.c, active)
		if i > 0 && active {
			previous := frontier[i-1]
			flow(img, previous.x+previous.w, previous.y+27, node.x, node.y+27, node.c, true)
		}
	}
	if frame > 8 {
		p := progress(frame, 8, 31)
		circle(img, moveInt(90, 820, p), moveInt(165, 365, p), 8, coral)
	}
	if frame > 25 {
		box(img, 356, 366, 612, 420, panelRaised, teal)
		drawText(img, 372, 378, "NEXT OPERATION GENERATED", 1, teal)
		drawText(img, 372, 398, "RESTORE EVIDENCE", 2, textPrimary)
		flow(img, 728, 325, 612, 393, teal, true)
	}
	drawText(img, 44, 424, "FOLLOW MINIMAL EDGE: LOCATE -> EXPLAIN -> CLASSIFY -> REPAIR", 1, textPrimary)
	caption(img, "UNKNOWN becomes actionable when the causal frontier expands to its blocker and next operation.")
}

func renderPrecedence(img *image.Paletted, _ manifestScene, frame int) {
	refutedP := progress(frame, 2, 11)
	unknownP := progress(frame, 8, 18)
	closedP := progress(frame, 14, 24)
	selectP := progress(frame, 20, 29)
	drawText(img, 40, 106, "PRECEDENCE STACK", 2, amber)
	drawText(img, 40, 134, "REFUTED > UNKNOWN > CLOSED", 1, textPrimary)
	stack := []struct {
		y     int
		title string
		lines []string
		c     uint8
		shape byte
		p     float64
	}{{172, "REFUTED", []string{"priority 130", "contradiction"}, coral, 'x', refutedP}, {258, "UNKNOWN", []string{"priority 60", "missing facts"}, unknown, 'd', unknownP}, {344, "CLOSED", []string{"priority 0", "no blocker"}, green, 's', closedP}}
	for i, card := range stack {
		stateCard(img, 54, card.y, 228, 70, card.title, card.lines, card.c, card.p > 0.5, card.shape)
		if card.p > 0.05 {
			circle(img, moveInt(312, 370, card.p), card.y+35, 8, card.c)
		}
		if i < 2 {
			line(img, 168, card.y+70, 168, stack[i+1].y, border)
		}
	}
	drawNode(img, 414, 198, 216, 142, "SELECT", []string{"highest priority", "fail closed"}, violet, selectP > 0.2)
	flow(img, 282, 207, 414, 240, coral, selectP > 0.1)
	flow(img, 630, 240, 704, 240, coral, selectP > 0.35)
	drawNode(img, 704, 198, 216, 90, "RESULT", []string{"REFUTED"}, coral, selectP > 0.45)
	if frame > 27 {
		box(img, 704, 318, 920, 414, panelRaised, amber)
		drawText(img, 718, 332, "LEDGER +1", 2, amber)
		drawText(img, 718, 364, "COUNTEREXAMPLE", 1, textPrimary)
		circle(img, moveInt(644, 718, progress(frame, 28, 35)), 366, 7, amber)
	}
	drawText(img, 416, 374, "COUNTEREXAMPLE APPENDS", 1, amber)
	drawText(img, 416, 398, "STATE IS NOT AVERAGED", 1, textPrimary)
	caption(img, "Candidates enter a precedence stack; REFUTED wins and its counterexample is appended to the ledger.")
}

func renderPackageResolution(img *image.Paletted, _ manifestScene, frame int) {
	sortP := progress(frame, 2, 17)
	mergeP := progress(frame, 15, 28)
	receiptP := progress(frame, 27, 35)
	activityY := moveInt(302, 144, sortP)
	entitiesY := moveInt(144, 302, sortP)
	drawText(img, 40, 104, "DISCOVERED FILES", 1, cyan)
	drawNode(img, 38, activityY, 218, 98, "activity.gooo", []string{"package billing", "PayOrder"}, cyan, frame > 2)
	drawNode(img, 38, entitiesY, 218, 98, "entities.gooo", []string{"package billing", "Order + Receipt"}, cyan, frame > 2)
	drawNode(img, 326, 192, 206, 164, "CANONICAL SORT", []string{"activity.gooo", "entities.gooo", "filename order"}, violet, sortP > 0.25)
	if sortP > 0.45 {
		circle(img, 310, 242, 8, violet)
		circle(img, 310, 302, 8, violet)
	}
	flow(img, 256, activityY+49, 326, 230, cyan, sortP > 0.2)
	flow(img, 256, entitiesY+49, 326, 318, cyan, sortP > 0.2)
	if mergeP > 0.2 {
		drawNode(img, 604, 192, 174, 164, "PACKAGE UNIT", []string{"parse", "merge", "one namespace"}, teal, true)
		flow(img, 532, 274, 604, 274, violet, true)
	}
	if receiptP > 0.2 {
		drawNode(img, 830, 192, 96, 164, "ENTRY", []string{"PayOrder", "receipt"}, amber, true)
		flow(img, 778, 274, 830, 274, teal, true)
	}
	if receiptP > 0.05 {
		circle(img, moveInt(540, 840, receiptP), 274, 8, amber)
	}
	drawText(img, 326, 382, "SORT -> PARSE EACH -> MERGE ONCE -> RESOLVE ENTRY", 1, textPrimary)
	drawText(img, 326, 408, "COLLISION / NAMESPACE MISMATCH = REJECT", 1, coral)
	caption(img, "Source cards physically reorder by canonical filename, merge into one package unit, then emit an entry receipt.")
}

func renderIncremental(img *image.Paletted, _ manifestScene, frame int) {
	selected := 1
	routeP := progress(frame, 2, 15)
	branchP := progress(frame, 12, 23)
	drawNode(img, 34, 194, 196, 124, "CHANGED NODE", []string{"surface", "six digests", "policy"}, cyan, frame > 2)
	flow(img, 230, 256, 286, 256, cyan, routeP > 0.15)
	diamond(img, 322, 256, 36, violet)
	drawText(img, 295, 238, "ROUTE", 1, dark)
	drawText(img, 296, 268, "BY ID", 1, dark)
	if routeP > 0.05 {
		circle(img, moveInt(238, 322, routeP), 256, 9, cyan)
	}
	ops := []struct {
		y     int
		title string
		lines []string
		c     uint8
		shape byte
	}{{108, "EXECUTE", []string{"fresh digest", "new receipt"}, teal, 's'}, {192, "REUSE", []string{"same digest", "exact replay"}, violet, 'o'}, {276, "UNKNOWN", []string{"oracle absent", "lower resolution"}, unknown, 'd'}, {360, "REFUTED", []string{"counterexample", "retain"}, coral, 'x'}}
	for i, op := range ops {
		active := branchP > 0.35 && selected == i
		stateCard(img, 480, op.y, 416, 68, op.title, op.lines, op.c, active, op.shape)
		flow(img, 358, 256, 480, op.y+34, op.c, active)
	}
	if branchP > 0.35 {
		drawText(img, 42, 374, "ONE IDENTITY", 1, amber)
		drawText(img, 42, 398, "ONE OPERATION RESULT", 1, textPrimary)
		drawText(img, 42, 422, "OTHER BRANCHES CLOSE", 1, textPrimary)
	}
	caption(img, "Six-digest identity routes a changed surface to exactly one terminal operation result.")
}

func renderBootstrap(img *image.Paletted, _ manifestScene, frame int) {
	laneP := progress(frame, 1, 14)
	compareP := progress(frame, 12, 25)
	gateP := progress(frame, 24, 34)
	drawNode(img, 32, 206, 144, 92, ".GOOO", []string{"bounded input"}, cyan, frame > 1)
	drawNode(img, 260, 132, 238, 136, "CANDIDATE KERNEL", []string{"bounded", "self-hosted", "semantic view"}, violet, laneP > 0.2)
	drawNode(img, 260, 326, 238, 136, "INDEPENDENT ORACLE", []string{"bootstrap", "separate facts", "witness"}, teal, laneP > 0.35)
	drawNode(img, 632, 196, 168, 168, "COMPARE", []string{"digest", "facts", "equal?"}, amber, compareP > 0.2)
	flow(img, 176, 252, 260, 200, cyan, laneP > 0.1)
	flow(img, 176, 252, 260, 394, cyan, laneP > 0.25)
	flow(img, 498, 200, 632, 244, violet, compareP > 0.1)
	flow(img, 498, 394, 632, 316, teal, compareP > 0.35)
	if laneP > 0.1 {
		circle(img, moveInt(176, 570, laneP), moveInt(252, 200, laneP), 8, violet)
	}
	if laneP > 0.25 {
		circle(img, moveInt(176, 570, laneP), moveInt(252, 394, laneP), 8, teal)
	}
	if compareP > 0.2 {
		circle(img, moveInt(570, 680, compareP), 276, 8, amber)
	}
	if compareP > 0.6 {
		stateCard(img, 800, 112, 124, 70, "EQUAL", []string{"parity"}, green, true, 's')
		flow(img, 800, 280, 800, 147, green, true)
	}
	stateCard(img, 800, 298, 124, 92, "REVIEW", []string{"NOT AUTO", "PROMOTED"}, amber, gateP > 0.2, 'd')
	if gateP > 0.2 {
		flow(img, 862, 182, 862, 298, amber, true)
	}
	drawText(img, 42, 382, "DIGEST + FACTS", 1, amber)
	drawText(img, 42, 406, "EQUAL OR MISMATCH -> REVIEW GATE", 1, textPrimary)
	caption(img, "Candidate and independent oracle facts compare at a terminal parity state, then stop at review.")
}

func renderPromotion(img *image.Paletted, _ manifestScene, frame int) {
	steps := []struct {
		x     int
		title string
		lines []string
		c     uint8
	}{{34, "FEATURE PR", []string{"artifact", "head SHA"}, cyan}, {218, "EXACT CI", []string{"exact head", "all checks"}, violet}, {402, "DEV", []string{"adopted", "receipt"}, teal}, {586, "OBSERVE", []string{"post-adoption", "fresh receipt"}, amber}, {770, "MAIN GATE", []string{"dev -> main", "promotion"}, green}}
	for i, step := range steps {
		active := frame >= 3+i*7
		drawNode(img, step.x, 202, 154, 118, step.title, step.lines, step.c, active)
		if i < 4 {
			flow(img, step.x+154, 261, steps[i+1].x, 261, step.c, active && frame >= 10+i*7)
		}
		if active {
			circle(img, step.x+20, 326, 7, step.c)
			drawText(img, step.x+34, 320, "RECEIPT", 1, step.c)
		}
	}
	finalRead := frame > 27
	box(img, 580, 130, 760, 178, panelRaised, amber)
	drawText(img, 592, 141, "FINAL EXACT HEAD", 1, amber)
	drawText(img, 592, 157, "REREAD", 1, amber)
	flow(img, 663, 202, 670, 178, amber, finalRead)
	if finalRead {
		circle(img, 670, 178, 7, amber)
		flow(img, 760, 154, 847, 202, amber, true)
	}
	if frame > 32 {
		box(img, 770, 342, 924, 396, panelRaised, green)
		drawText(img, 783, 354, "GATE OPEN", 1, green)
		circle(img, moveInt(742, 790, progress(frame, 28, 35)), 261, 7, amber)
	} else if frame > 25 {
		box(img, 770, 342, 924, 396, panelRaised, amber)
		drawText(img, 783, 354, "FRESH RECEIPT", 1, amber)
	} else {
		box(img, 770, 342, 924, 396, panel, coral)
		cross(img, 782, 348, 812, 378, coral)
		drawText(img, 820, 352, "STOP", 1, coral)
		drawText(img, 783, 370, "MISSING RECEIPT", 1, coral)
	}
	drawText(img, 38, 378, "RECEIPTS IN ORDER", 1, amber)
	drawText(img, 38, 402, "FEATURE -> CI -> DEV -> OBSERVE -> MAIN", 1, textPrimary)
	caption(img, "Each receipt unlocks the next gate; main opens only after exact-head CI and post-adoption observation.")
}
