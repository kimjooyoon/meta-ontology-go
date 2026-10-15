package languagesyntax

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestDeliveryOriginSourceLawsFresh1615(t *testing.T) {
	path := "internal/meta/selfimprovementtransport/originpolicy/source.gooo"
	result := replay.ExecuteWithOrderedCollections(os.DirFS("../../../.."), path, KindValid, "")
	if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
		!result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("expanded delivery origin source lost preservation laws: %+v", result)
	}
}
