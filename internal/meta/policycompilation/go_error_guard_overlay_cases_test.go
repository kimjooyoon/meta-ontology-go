package policycompilation

import (
	"bytes"
	"testing"
)

const goGuardOverlayCurrent = "package main\nimport \"fmt\"\nfunc subject(value int) int { fmt.Println(value); return value }\nfunc sibling() int { return 7 }\n"
const goGuardOverlayVariant = "package main\nimport \"fmt\"\nfunc subject(value int) int { fmt.Println(value); return value + 1 }\nfunc historicalSibling() int { return 99 }\n"

func TestGoErrorGuardOverlayFollowsDeclarationWithoutReintroducingSiblings(t *testing.T) {
	for _, owner := range []string{"original.go", "generated-part-17.go"} {
		t.Run(owner, func(t *testing.T) {
			files := map[string][]byte{
				"empty.go":        []byte("package main\n"),
				owner:             []byte(goGuardOverlayCurrent),
				"frozen_test.go":  []byte("package main\n// frozen oracle\n"),
				"other-helper.go": []byte("package main\nfunc unrelated() int { return 3 }\n"),
			}
			original, oracle := bytes.Clone(files[owner]), bytes.Clone(files["frozen_test.go"])
			path, rebound, err := bindGoGuardDeclaration(files, []string{"empty.go", owner, "other-helper.go"}, "subject", []byte(goGuardOverlayVariant))
			if err != nil || path != owner {
				t.Fatalf("declaration binding = %s: %v", path, err)
			}
			if !bytes.Contains(rebound, []byte("return value + 1")) ||
				!bytes.Contains(rebound, []byte("func sibling() int { return 7 }")) ||
				bytes.Contains(rebound, []byte("historicalSibling")) || bytes.Count(rebound, []byte("func subject(")) != 1 ||
				!bytes.Equal(original, files[owner]) || !bytes.Equal(oracle, files["frozen_test.go"]) {
				t.Fatal("declaration overlay changed a sibling, duplicated a declaration, or changed the frozen input")
			}
		})
	}
}

func TestGoErrorGuardOverlayRejectsMissingAmbiguousAndChangedContext(t *testing.T) {
	for _, name := range []string{"missing", "ambiguous", "signature", "import", "package", "syntax"} {
		t.Run(name, func(t *testing.T) {
			files := map[string][]byte{"current.go": []byte(goGuardOverlayCurrent)}
			production := []string{"current.go"}
			switch name {
			case "missing":
				files["current.go"] = []byte("package main\n")
			case "ambiguous":
				files["second.go"] = []byte(goGuardOverlayCurrent)
				production = append(production, "second.go")
			case "signature":
				files["current.go"] = bytes.ReplaceAll(files["current.go"], []byte("value int"), []byte("value int64"))
			case "import":
				files["current.go"] = bytes.ReplaceAll(files["current.go"], []byte("import \"fmt\""), []byte("import fmt \"example.invalid/notfmt\""))
			case "package":
				files["current.go"] = bytes.ReplaceAll(files["current.go"], []byte("package main"), []byte("package other"))
			case "syntax":
				files["current.go"] = []byte("not valid Go")
			}
			path, rebound, err := bindGoGuardDeclaration(files, production, "subject", []byte(goGuardOverlayVariant))
			if err == nil || path != "" || rebound != nil {
				t.Fatalf("unsupported context produced an overlay: %s err=%v", path, err)
			}
		})
	}
}
