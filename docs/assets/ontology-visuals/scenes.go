package main

import "image"

type sceneRenderer func(*image.Paletted, manifestScene, int)

var renderers = map[string]sceneRenderer{
	"ontology.visual.intent-ir-lowering":        renderSourceToGo,
	"ontology.visual.authority-boundary":        renderSemanticIR,
	"ontology.visual.munchausen-proof-choice":   renderMultifilePackage,
	"ontology.visual.claim-evidence-lifecycle":  renderAgentHandoff,
	"ontology.visual.unknown-cause-descent":     renderUnknownResolution,
	"ontology.visual.precedence-counterexample": renderRefutationPrecedence,
	"ontology.visual.package-resolution":        renderDeterministicReplay,
	"ontology.visual.incremental-conformance":   renderIncrementalConformance,
	"ontology.visual.bootstrap-oracle":           renderSelfImprovement,
	"ontology.visual.promotion-lineage":         renderDomainProjection,
}

func renderFrame(scene manifestScene, frame int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, 0, 0, width, height, background)
	drawText(img, 28, 20, "GOOO ENGINEERING NARRATIVES", 2, textPrimary)
	drawText(img, 30, 47, scene.Title, 1, cyan)
	fill(img, 742, 18, 932, 50, panelRaised)
	stroke(img, 742, 18, 932, 50, violet)
	drawText(img, 756, 29, fmtSceneNumber(scene.Sequence), 1, violet)
	renderers[scene.ID](img, scene, frame)
	fill(img, 28, 480, 932, 512, panelRaised)
	drawText(img, 42, 488, "TEN STORIES / TEN TERMINALS", 1, textMuted)
	drawText(img, 302, 488, "GENERATED FROM STORY MANIFEST", 1, textMuted)
	drawText(img, 682, 488, "NO AGGREGATE SCORE", 1, amber)
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

func renderSourceToGo(img *image.Paletted, _ manifestScene, frame int) {
	pIR := progress(frame, 5, 17)
	pGo := progress(frame, 15, 28)
	pReceipt := progress(frame, 26, 35)
	codeBlock(img, 32, 150, 260, 132, "INPUT: ACTIVITY.GOOO", []string{"activity PayOrder(", "  Order) -> Receipt"}, cyan, frame > 1)
	codeBlock(img, 350, 150, 246, 132, "LOWERED: SEMANTIC IR", []string{"activity=PayOrder", "in=Order", "out=Receipt"}, violet, pIR > 0.2)
	codeBlock(img, 658, 112, 264, 150, "OUTPUT: GENERATED GO", []string{"// generated", "func PayOrder(", "  Order) Receipt {"}, teal, pGo > 0.2)
	codeBlock(img, 658, 292, 264, 104, "RECEIPT.JSON", []string{"decision: PASS", "kind: interface", "writes: 0"}, amber, pReceipt > 0.2)
	flow(img, 292, 216, 350, 216, cyan, pIR > 0.1)
	flow(img, 596, 216, 658, 188, violet, pGo > 0.1)
	flow(img, 596, 216, 658, 344, violet, pReceipt > 0.1)
	if pIR > 0.05 {
		labeledToken(img, moveInt(292, 488, pIR), 216, "IR", violet)
	}
	if pGo > 0.05 {
		labeledToken(img, moveInt(596, 724, pGo), moveInt(216, 188, pGo), "GO", teal)
	}
	if pReceipt > 0.05 {
		labeledToken(img, moveInt(596, 724, pReceipt), moveInt(216, 344, pReceipt), "EVIDENCE", amber)
	}
	drawText(img, 42, 356, "PARSE", 1, cyan)
	drawText(img, 140, 356, "->", 1, textMuted)
	drawText(img, 174, 356, "LOWER", 1, violet)
	drawText(img, 270, 356, "->", 1, textMuted)
	drawText(img, 304, 356, "EMIT", 1, teal)
	drawText(img, 390, 356, "->", 1, textMuted)
	drawText(img, 424, 356, "CONSUME", 1, amber)
	drawText(img, 42, 388, "SOURCE ARTIFACT", 1, cyan)
	drawText(img, 210, 388, "BECOMES", 1, textPrimary)
	drawText(img, 304, 388, "GO + RECEIPT", 2, teal)
	caption(img, "A concrete billing declaration is parsed, lowered, and consumed as generated Go plus a PASS receipt.")
}

func renderSemanticIR(img *image.Paletted, _ manifestScene, frame int) {
	pNode := progress(frame, 3, 16)
	pEdge := progress(frame, 12, 24)
	pBackend := progress(frame, 22, 34)
	codeBlock(img, 32, 178, 246, 136, "DECLARATION", []string{"entity Order", "activity PayOrder", "-> Receipt"}, cyan, frame > 2)
	box(img, 346, 120, 630, 356, panel, violet)
	drawText(img, 364, 134, "INSPECTABLE SEMANTIC IR", 1, violet)
	circle(img, 440, 212, 18, cyan)
	drawText(img, 468, 204, "Order", 2, textPrimary)
	circle(img, 440, 314, 18, teal)
	drawText(img, 468, 306, "PayOrder", 2, textPrimary)
	circle(img, 440, 416, 18, amber)
	drawText(img, 468, 408, "Receipt", 2, textPrimary)
	if pNode > 0.2 {
		line(img, 440, 230, 440, 296, cyan)
		line(img, 440, 332, 440, 398, teal)
	}
	if pEdge > 0.2 {
		arrow(img, 440, 296, cyan)
		arrow(img, 440, 398, teal)
	}
	drawText(img, 548, 244, "used", 1, cyan)
	drawText(img, 548, 346, "wasGeneratedBy", 1, teal)
	codeBlock(img, 690, 172, 230, 142, "BACKEND", []string{"consume(IR)", "resolve IDs", "emit source map"}, teal, pBackend > 0.2)
	flow(img, 630, 314, 690, 244, violet, pBackend > 0.1)
	if pBackend > 0.1 {
		labeledToken(img, moveInt(630, 690, pBackend), moveInt(314, 244, pBackend), "GRAPH", violet)
	}
	drawText(img, 42, 360, "DECLARATION", 1, cyan)
	drawText(img, 190, 360, "-> TYPED NODES + EDGES", 1, violet)
	drawText(img, 42, 390, "IR IS THE INTERCHANGE; BACKEND READS IT", 1, textPrimary)
	caption(img, "One declaration becomes typed semantic nodes and edges; the backend consumes that graph as its input artifact.")
}

func renderMultifilePackage(img *image.Paletted, _ manifestScene, frame int) {
	sortP := progress(frame, 2, 16)
	mergeP := progress(frame, 14, 26)
	apiP := progress(frame, 24, 35)
	activityY := moveInt(304, 142, sortP)
	entitiesY := moveInt(142, 304, sortP)
	codeBlock(img, 34, activityY, 238, 96, "activity.gooo", []string{"PayOrder(Order)", "-> Receipt"}, cyan, frame > 1)
	codeBlock(img, 34, entitiesY, 238, 96, "entities.gooo", []string{"entity Order", "entity Receipt"}, cyan, frame > 1)
	box(img, 334, 130, 560, 358, panel, violet)
	drawText(img, 352, 144, "CANONICAL FILE ORDER", 1, violet)
	codeBlock(img, 354, 180, 188, 74, "01", []string{"activity.gooo"}, violet, sortP > 0.4)
	codeBlock(img, 354, 282, 188, 74, "02", []string{"entities.gooo"}, violet, sortP > 0.4)
	if sortP > 0.15 {
		circle(img, moveInt(278, 344, sortP), moveInt(activityY+48, 217, sortP), 8, violet)
		circle(img, moveInt(278, 344, sortP), moveInt(entitiesY+48, 319, sortP), 8, violet)
	}
	flow(img, 272, activityY+48, 334, 217, cyan, sortP > 0.1)
	flow(img, 272, entitiesY+48, 334, 319, cyan, sortP > 0.1)
	if mergeP > 0.2 {
		codeBlock(img, 608, 176, 174, 180, "PACKAGE API", []string{"package billing", "func PayOrder(", "  Order) Receipt"}, teal, true)
		flow(img, 560, 268, 608, 268, violet, true)
	}
	if apiP > 0.2 {
		codeBlock(img, 814, 176, 112, 180, "RECEIPT", []string{"PASS", "entry", "PayOrder"}, amber, true)
		flow(img, 782, 268, 814, 268, teal, true)
	}
	drawText(img, 350, 390, "SORT", 1, violet)
	drawText(img, 424, 390, "->", 1, textMuted)
	drawText(img, 458, 390, "PARSE EACH", 1, cyan)
	drawText(img, 578, 390, "->", 1, textMuted)
	drawText(img, 612, 390, "MERGE", 1, teal)
	drawText(img, 704, 390, "->", 1, textMuted)
	drawText(img, 738, 390, "ENTRY", 1, amber)
	caption(img, "Two source cards physically reorder by filename, merge into one package API, and produce an entry receipt.")
}

func renderAgentHandoff(img *image.Paletted, _ manifestScene, frame int) {
	authorP := progress(frame, 2, 12)
	compilerP := progress(frame, 10, 23)
	reviewerP := progress(frame, 21, 34)
	lane(img, 140, "AUTHOR AGENT / INTENT OWNER", cyan)
	lane(img, 270, "COMPILER AGENT / OUTPUT WRITER", violet)
	lane(img, 400, "REVIEWER AGENT / RECEIPT CONSUMER", amber)
	codeBlock(img, 42, 154, 250, 92, "WRITES", []string{"main.gooo", "PayOrder(Order)"}, cyan, authorP > 0.2)
	codeBlock(img, 354, 284, 250, 92, "CREATES", []string{"generated.go", "// generated", "func PayOrder"}, violet, compilerP > 0.2)
	codeBlock(img, 674, 414-80, 250, 92, "READS", []string{"receipt.json", "subject SHA", "evidence digest"}, amber, reviewerP > 0.2)
	line(img, 310, 200, 338, 330, cyan)
	line(img, 622, 330, 658, 380, violet)
	if authorP > 0.05 {
		labeledToken(img, moveInt(292, 354, authorP), moveInt(200, 330, authorP), "INTENT", cyan)
	}
	if compilerP > 0.05 {
		labeledToken(img, moveInt(604, 674, compilerP), moveInt(330, 380, compilerP), "GO", violet)
	}
	if reviewerP > 0.05 {
		labeledToken(img, moveInt(658, 782, reviewerP), moveInt(380, 380, reviewerP), "RECEIPT", amber)
	}
	fill(img, 306, 154, 322, 246, panelRaised)
	stroke(img, 306, 154, 322, 246, coral)
	drawText(img, 308, 174, "NO", 1, coral)
	drawText(img, 308, 192, "CROSS", 1, coral)
	drawText(img, 308, 210, "WRITE", 1, coral)
	drawText(img, 42, 438, "EACH AGENT HAS ONE WRITE SET", 1, textPrimary)
	caption(img, "The author writes intent, the compiler creates caller-owned output, and the reviewer consumes the immutable receipt.")
}

func renderUnknownResolution(img *image.Paletted, _ manifestScene, frame int) {
	fieldsP := progress(frame, 2, 18)
	resolverP := progress(frame, 16, 27)
	evidenceP := progress(frame, 25, 35)
	codeBlock(img, 34, 146, 254, 154, "MISSING ARTIFACT", []string{"claim_id: C-17", "artifact: absent", "decision: UNKNOWN"}, unknown, frame > 1)
	box(img, 326, 112, 610, 356, panel, amber)
	drawText(img, 344, 126, "UNKNOWN FIELDS", 1, amber)
	fields := []struct{ name, value string; c uint8 }{{"stage", "verify", cyan}, {"step", "observe", violet}, {"reason", "absent", amber}, {"class", "direct", unknown}, {"next_operation", "restore", teal}, {"blocked_by", "artifact", coral}}
	for i, f := range fields {
		y := 164 + i*42
		active := fieldsP > float64(i)/6
		smallNode(img, 348, y, 248, 32, f.name, f.value, f.c, active)
		if i < 5 && active {
			flow(img, 596, y+16, 620, y+58, f.c, true)
		}
	}
	codeBlock(img, 654, 170, 256, 106, "RESOLVER", []string{"next_operation()", "restore evidence"}, teal, resolverP > 0.2)
	codeBlock(img, 654, 306, 256, 104, "RE-EVALUATE C-17", []string{"evidence: present", "decision: CLOSED"}, green, evidenceP > 0.3)
	flow(img, 610, 334, 654, 222, teal, resolverP > 0.2)
	flow(img, 782, 276, 782, 306, green, evidenceP > 0.2)
	if evidenceP > 0.05 {
		labeledToken(img, moveInt(590, 782, evidenceP), moveInt(334, 306, evidenceP), "EVIDENCE", green)
	}
	drawText(img, 40, 344, "UNKNOWN", 1, unknown)
	drawText(img, 126, 344, "-> SIX FIELDS", 1, textPrimary)
	drawText(img, 40, 374, "-> RESOLVE -> SAME CLAIM ID -> CLOSED", 1, green)
	caption(img, "Missing evidence first creates six actionable fields; a resolver supplies evidence before the same claim closes.")
}

func renderRefutationPrecedence(img *image.Paletted, _ manifestScene, frame int) {
	pUnknown := progress(frame, 2, 12)
	pRefuted := progress(frame, 9, 20)
	pSelect := progress(frame, 18, 28)
	pLedger := progress(frame, 26, 35)
	codeBlock(img, 34, 158, 270, 126, "INCOMPLETE EVIDENCE", []string{"claim: C-17", "evidence: partial", "state: UNKNOWN"}, unknown, pUnknown > 0.2)
	codeBlock(img, 34, 312, 270, 112, "COUNTEREXAMPLE", []string{"same claim: C-17", "clock order: invalid", "state: REFUTED"}, coral, pRefuted > 0.2)
	box(img, 350, 120, 608, 424, panel, violet)
	drawText(img, 370, 136, "PRECEDENCE", 1, amber)
	drawText(img, 370, 160, "REFUTED > UNKNOWN > CLOSED", 1, textPrimary)
	stateCard(img, 382, 198, 234, 74, "REFUTED", []string{"priority 130"}, coral, pRefuted > 0.5, 'x')
	stateCard(img, 382, 298, 234, 74, "UNKNOWN", []string{"priority 60"}, unknown, pUnknown > 0.5 && pRefuted < 0.5, 'd')
	stateCard(img, 382, 398, 234, 74, "CLOSED", []string{"priority 0"}, green, false, 's')
	flow(img, 304, 370, 382, 235, coral, pSelect > 0.1)
	if pSelect > 0.05 {
		labeledToken(img, moveInt(304, 382, pSelect), 370, "SELECT", coral)
	}
	codeBlock(img, 674, 198, 238, 104, "TERMINAL", []string{"decision: REFUTED", "adoption: blocked"}, coral, pSelect > 0.5)
	codeBlock(img, 674, 338, 238, 104, "APPEND-ONLY LEDGER", []string{"counterexample: C-17", "record: retained"}, amber, pLedger > 0.3)
	flow(img, 616, 235, 674, 250, coral, pSelect > 0.3)
	flow(img, 616, 435, 674, 390, amber, pLedger > 0.3)
	caption(img, "A counterexample arrives after incomplete evidence; precedence selects REFUTED and appends the record unchanged.")
}

func renderDeterministicReplay(img *image.Paletted, _ manifestScene, frame int) {
	pInputs := progress(frame, 2, 12)
	pRun := progress(frame, 10, 22)
	pChange := progress(frame, 21, 29)
	pBlock := progress(frame, 28, 35)
	codeBlock(img, 32, 150, 254, 142, "PINNED INPUTS", []string{"source sha: aaa", "contract sha: bbb", "toolchain: ccc"}, cyan, pInputs > 0.2)
	codeBlock(img, 354, 126, 238, 132, "RUN 1", []string{"output.go sha: 111", "receipt sha: 222"}, violet, pRun > 0.2)
	codeBlock(img, 354, 292, 238, 132, "RUN 2", []string{"output.go sha: 111", "receipt sha: 222"}, teal, pRun > 0.5)
	box(img, 646, 126, 926, 424, panel, amber)
	drawText(img, 666, 142, "DIGEST COMPARATOR", 1, amber)
	drawText(img, 666, 174, "SOURCE + CONTRACT + TOOLCHAIN", 1, textPrimary)
	if pRun > 0.2 {
		flow(img, 592, 192, 646, 192, violet, true)
		flow(img, 592, 358, 646, 358, teal, true)
	}
	if pChange < 0.5 {
		stateCard(img, 666, 214, 220, 74, "BYTE IDENTICAL", []string{"replay_count: 2"}, green, pRun > 0.6, 's')
	} else {
		stateCard(img, 666, 214, 220, 74, "MISMATCH", []string{"byte 37 changed"}, coral, true, 'x')
	}
	if pChange > 0.2 {
		codeBlock(img, 666, 314, 220, 80, "CHANGED BYTE", []string{"output.go[37]: +1"}, coral, true)
	}
	if pBlock > 0.2 {
		stateCard(img, 666, 400, 220, 70, "BLOCKED", []string{"adoption mismatch"}, coral, true, 'd')
		flow(img, 776, 394, 776, 400, coral, true)
	}
	if pChange > 0.05 {
		labeledToken(img, moveInt(592, 666, pChange), 254, "BYTE", coral)
	}
	caption(img, "Pinned digests replay byte-identically first; a causal byte change becomes a mismatch and blocks adoption.")
}

func renderIncrementalConformance(img *image.Paletted, _ manifestScene, frame int) {
	pIdentity := progress(frame, 2, 13)
	pRoute := progress(frame, 11, 23)
	codeBlock(img, 32, 144, 258, 142, "CHANGED SURFACE", []string{"subject: C-17", "source sha: aaa", "contract sha: bbb"}, cyan, pIdentity > 0.2)
	box(img, 332, 126, 590, 430, panel, violet)
	drawText(img, 350, 142, "SIX-DIGEST IDENTITY", 1, violet)
	codeBlock(img, 350, 174, 224, 102, "REPLAY RECEIPT", []string{"replay_count: 2", "deterministic: true"}, teal, pIdentity > 0.4)
	diamond(img, 620, 246, 38, amber)
	drawText(img, 594, 236, "ROUTE", 1, dark)
	drawText(img, 596, 254, "CHECK", 1, dark)
	flow(img, 290, 214, 582, 246, cyan, pRoute > 0.1)
	if pRoute > 0.05 {
		labeledToken(img, moveInt(290, 620, pRoute), moveInt(214, 246, pRoute), "IDENTITY", cyan)
	}
	stateCard(img, 694, 140, 210, 68, "EXECUTE", []string{"fresh receipt"}, teal, false, 's')
	stateCard(img, 694, 220, 210, 68, "REUSE", []string{"receipt replay"}, violet, pRoute > 0.55, 'o')
	stateCard(img, 694, 300, 210, 68, "UNKNOWN", []string{"lower resolution"}, unknown, false, 'd')
	stateCard(img, 694, 380, 210, 68, "REFUTED", []string{"retain blocker"}, coral, false, 'x')
	flow(img, 658, 246, 694, 254, violet, pRoute > 0.5)
	if pRoute > 0.5 {
		codeBlock(img, 350, 324, 224, 100, "SELECTED JOB", []string{"REUSE", "receipt: 222", "count: 2"}, violet, true)
	}
	drawText(img, 42, 346, "AFFECTED CHECK", 1, amber)
	drawText(img, 182, 346, "-> REUSE RECEIPT", 1, violet)
	drawText(img, 42, 378, "OTHER OUTCOMES REMAIN STATIC LEGEND", 1, textPrimary)
	caption(img, "A changed subject is identified by six digests and routes to one reused receipt while other outcomes remain inactive.")
}

func renderSelfImprovement(img *image.Paletted, _ manifestScene, frame int) {
	pRule := progress(frame, 2, 12)
	pCI := progress(frame, 10, 21)
	pDev := progress(frame, 19, 29)
	pMain := progress(frame, 27, 35)
	lane(img, 116, "OBSERVATION / METRIC OWNER", cyan)
	lane(img, 230, "META-RULE / EVALUATOR", violet)
	lane(img, 344, "EXACT-HEAD CI + DEV", teal)
	lane(img, 416, "PROMOTION ELIGIBILITY", amber)
	codeBlock(img, 42, 132, 230, 72, "OBSERVED BUG", []string{"metric drift", "evidence missing"}, cyan, pRule > 0.15)
	codeBlock(img, 330, 246, 238, 72, "RULE CHANGE", []string{"fail-closed", "new evaluator"}, violet, pRule > 0.45)
	codeBlock(img, 620, 132, 270, 72, "EXACT-HEAD CI", []string{"head SHA pinned", "all checks"}, teal, pCI > 0.2)
	codeBlock(img, 42, 360, 230, 72, "DEV", []string{"adopted", "receipt"}, teal, pDev > 0.2)
	codeBlock(img, 330, 360, 238, 72, "POST-ADOPTION", []string{"observe", "fresh receipt"}, amber, pDev > 0.55)
	codeBlock(img, 620, 360, 270, 72, "MAIN ELIGIBLE", []string{"only after all", "gates closed"}, green, pMain > 0.3)
	flow(img, 272, 168, 330, 282, cyan, pRule > 0.2)
	flow(img, 568, 282, 620, 168, violet, pCI > 0.2)
	flow(img, 755, 204, 755, 360, teal, pDev > 0.3)
	flow(img, 272, 396, 330, 396, teal, pDev > 0.5)
	flow(img, 568, 396, 620, 396, amber, pMain > 0.2)
	if pRule > 0.05 {
		labeledToken(img, moveInt(272, 330, pRule), moveInt(168, 282, pRule), "RULE", violet)
	}
	if pCI > 0.05 {
		labeledToken(img, moveInt(568, 755, pCI), moveInt(282, 168, pCI), "CI", teal)
	}
	caption(img, "An observed bug changes a meta-rule, then exact-head CI, dev adoption, and post-adoption evidence unlock eligibility in order.")
}

func renderDomainProjection(img *image.Paletted, _ manifestScene, frame int) {
	pInput := progress(frame, 2, 12)
	pContract := progress(frame, 10, 22)
	pDossier := progress(frame, 20, 31)
	drawText(img, 40, 100, "EXPERIMENTAL / NOT YET CLOSED", 2, coral)
	drawText(img, 42, 128, "PROPOSED DOMAIN PROJECTION", 1, amber)
	codeBlock(img, 32, 164, 258, 126, "SERVICE CONTRACT", []string{"OpenAPI: /orders", "status: 200", "schema: Receipt"}, cyan, pInput > 0.2)
	codeBlock(img, 32, 316, 258, 112, "INFRA FACTS", []string{"tofu show -json", "plan: present", "apply: forbidden"}, teal, pInput > 0.4)
	box(img, 352, 148, 598, 426, panel, violet)
	drawText(img, 372, 164, "PROPOSED GOOO CONTRACT", 1, violet)
	drawText(img, 372, 192, "service Receipt", 2, cyan)
	drawText(img, 372, 226, "bind", 1, amber)
	drawText(img, 372, 258, "infra Receipt", 2, teal)
	line(img, 520, 208, 520, 244, amber)
	arrow(img, 520, 244, amber)
	if pContract > 0.2 {
		codeBlock(img, 640, 198, 264, 116, "MISMATCH DOSSIER", []string{"service: Receipt.v2", "infra: Receipt.v1", "diff: schema"}, coral, true)
		flow(img, 598, 258, 640, 258, violet, true)
	}
	if pDossier > 0.2 {
		stateCard(img, 640, 342, 264, 74, "UNKNOWN", []string{"implementation absent"}, unknown, true, 'd')
		drawText(img, 372, 376, "NO INFRA MUTATION", 1, coral)
		drawText(img, 372, 400, "NO PROMOTION", 1, coral)
		flow(img, 772, 314, 772, 342, unknown, true)
	}
	if pDossier > 0.05 {
		labeledToken(img, moveInt(598, 772, pDossier), 258, "DOSSIER", coral)
	}
	caption(img, "EXPERIMENTAL: a proposed service/infrastructure projection emits a mismatch dossier and stays UNKNOWN until evidence exists.")
}
