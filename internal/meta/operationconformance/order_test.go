package operationconformance

import "testing"

func TestDuplicateInitBodiesUseDistinctOriginalOrdinals(t *testing.T) {
	source := FileEvidence{Path: "duplicate-init.go", Data: []byte("package fixture\n\nfunc init() { println(1) }\n\nfunc init() { println(1) }\n")}
	orders, err := declarationOrders(source)
	if err != nil {
		t.Fatal(err)
	}
	candidates := []FileEvidence{
		{Path: "duplicate-init_split01.go", Data: []byte("package fixture\n\nfunc init() { println(1) }\n"), DeclarationOrder: []DeclarationOrder{orders[0]}},
		{Path: "duplicate-init_split02.go", Data: []byte("package fixture\n\nfunc init() { println(1) }\n"), DeclarationOrder: []DeclarationOrder{orders[1]}},
	}
	evidence := SplitGoEvidence{Source: source, Candidates: candidates}
	if observeOrder(evidence) != DecisionPass {
		t.Fatal("distinct original ordinals for duplicate init bodies were rejected")
	}

	candidates[1].DeclarationOrder[0].Ordinal = candidates[0].DeclarationOrder[0].Ordinal
	evidence.Candidates = candidates
	if observeOrder(evidence) != DecisionFail {
		t.Fatal("duplicate original ordinal was accepted")
	}
}

func TestNormalDeclarationReorderingDoesNotChangeInitializationOrder(t *testing.T) {
	source := FileEvidence{Path: "normal.go", Data: []byte("package fixture\n\nfunc first() {}\n\nfunc second() {}\n")}
	candidates := []FileEvidence{
		{Path: "normal_split01.go", Data: []byte("package fixture\n\nfunc second() {}\n")},
		{Path: "normal_split02.go", Data: []byte("package fixture\n\nfunc first() {}\n")},
	}
	if observeOrder(SplitGoEvidence{Source: source, Candidates: candidates}) != DecisionPass {
		t.Fatal("ordinary function reordering changed initialization order")
	}
}

func TestInitializationUnitReorderingIsRejected(t *testing.T) {
	source := FileEvidence{Path: "init.go", Data: []byte("package fixture\n\nvar first = len(\"first\")\n\nvar second = len(\"second\")\n")}
	candidates := []FileEvidence{
		{Path: "init_split01.go", Data: []byte("package fixture\n\nvar second = len(\"second\")\n")},
		{Path: "init_split02.go", Data: []byte("package fixture\n\nvar first = len(\"first\")\n")},
	}
	if observeOrder(SplitGoEvidence{Source: source, Candidates: candidates}) != DecisionFail {
		t.Fatal("initialization unit reordering was accepted")
	}
}
