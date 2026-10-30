package bodylibrary

import (
	"context"
	"strings"
	"testing"
)

func TestLibraryEntryUsesStableActivityIdentityAcrossDisplayRename(t *testing.T) {
	const source = `package releaseidentity
namespace releaseidentity
entity Text id "urn:gooo:type:string"
activity CurrentVersion() -> Text id "release://activity/current-version" computes "return \"0.6.39-dev\""
activity Unrelated() -> Text id "release://activity/unrelated" computes "return \"wrong\""
`
	for _, name := range []string{"CurrentVersion", "RenamedVersion"} {
		t.Run(name, func(t *testing.T) {
			current := strings.ReplaceAll(source, "CurrentVersion", name)
			generated, err := Generate(context.Background(), "version.gooo", []byte(current),
				"release://activity/current-version", "releaseidentity")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(generated, `return "0.6.39-dev"`) ||
				strings.Contains(generated, `return "wrong"`) ||
				!strings.Contains(generated, `Evaluate returns activity="release://activity/current-version"`) {
				t.Fatal("library entry lost the selected source identity", generated)
			}
		})
	}
}

func TestLibraryEntryRefusesUnknownIdentity(t *testing.T) {
	const source = `package collision
namespace collision
entity Text id "urn:gooo:type:string"
activity First() -> Text id "urn:first" computes "return \"first\""
activity Other() -> Text id "urn:other" computes "return \"second\""
`
	if _, err := Generate(context.Background(), "collision.gooo", []byte(source), "First", "collision"); err != nil {
		t.Fatal("control source must be valid", err)
	}
	if _, err := Generate(context.Background(), "collision.gooo", []byte(source), "urn:missing", "collision"); err == nil || !strings.Contains(err.Error(), "one activity") {
		t.Fatal("unknown identity must refuse", err)
	}
}
