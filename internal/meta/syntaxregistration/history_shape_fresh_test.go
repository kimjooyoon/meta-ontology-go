package syntaxregistration

import (
	"reflect"
	"strings"
	"testing"
)

func TestHistorySelectorUsesSyntaxInsteadOfLineLayout(t *testing.T) {
	source := parseHistoryCapacityFixture(t, historyCapacityLegacyFixture())
	history, err := readHistoryOrigins(source)
	if err != nil {
		t.Fatal(err)
	}
	writeHistoryOrigins(source, history)
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	raw := string(changes[closureRoot+"history.go"])
	compact := strings.Replace(raw, historySelector, "if target,known:=historicalSyntaxTargetChunk0(version);known{expected[0].Target=target}", 1)
	compact = strings.ReplaceAll(compact, "return 0, false", "return 0,false")
	if compact == raw {
		t.Fatal("no presentation change")
	}
	observed, err := readHistoryOrigins(parseHistoryCapacityFixture(t, []byte(compact)))
	if err != nil || !reflect.DeepEqual(history.targets, observed.targets) {
		t.Fatal("line layout changed historical meaning", err)
	}
	for _, bad := range []string{
		strings.Replace(raw, "func historicalSyntaxTargetChunk0(version int)", "func historicalSyntaxTargetChunk0(other int)", 1),
		strings.Replace(raw, "func historicalSyntaxTargetChunk0(version int)", "func historicalSyntaxTargetChunk0()", 1),
		strings.Replace(raw, "(int, bool)", "(int, string)", 1),
		strings.Replace(raw, "(int, bool)", "", 1),
	} {
		if bad == raw {
			t.Fatal("signature mutation missing")
		}
		if _, err := readHistoryOrigins(parseHistoryCapacityFixture(t, []byte(bad))); err == nil {
			t.Fatal("wrong helper signature accepted")
		}
	}
}
