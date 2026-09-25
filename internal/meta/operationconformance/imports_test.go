package operationconformance

import "testing"

func TestImportIdentityUsesDeduplicatedSetUnion(t *testing.T) {
	source := FileEvidence{Path: "source.go", Data: []byte("package fixture\n\nimport (\n\t\"errors\"\n\t\"fmt\"\n)\n\nfunc first() { fmt.Println(errors.New(\"x\")) }\n")}
	candidates := []FileEvidence{
		{Path: "part01.go", Data: []byte("package fixture\n\nimport \"fmt\"\n\nfunc first() { fmt.Println(\"x\") }\n")},
		{Path: "part02.go", Data: []byte("package fixture\n\nimport (\n\t\"errors\"\n\t\"fmt\"\n)\n\nfunc second() { fmt.Println(errors.New(\"y\")) }\n")},
	}
	if decision := observeImports(SplitGoEvidence{Source: source, Candidates: candidates}); decision != DecisionPass {
		t.Fatalf("duplicate import set was rejected: %s", decision)
	}

	mutations := []struct {
		name string
		data []byte
	}{
		{name: "missing", data: []byte("package fixture\n\nimport \"fmt\"\n\nfunc second() { fmt.Println(\"y\") }\n")},
		{name: "alias", data: []byte("package fixture\n\nimport (\n\te \"errors\"\n\t\"fmt\"\n)\n\nfunc second() { fmt.Println(e.New(\"y\")) }\n")},
		{name: "path", data: []byte("package fixture\n\nimport (\n\t\"errors\"\n\tstrings \"strings\"\n)\n\nfunc second() { fmt.Println(strings.TrimSpace(\"y\")) }\n")},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := append([]FileEvidence{}, candidates...)
			changed[1].Data = mutation.data
			if observeImports(SplitGoEvidence{Source: source, Candidates: changed}) != DecisionFail {
				t.Fatalf("%s import drift was accepted", mutation.name)
			}
		})
	}
}
