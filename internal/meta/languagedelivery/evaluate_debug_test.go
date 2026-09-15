package languagedelivery

import "testing"

func TestCanonicalDebuggerUsesExternalDebugReceipt(t *testing.T) {
	contract := CanonicalContract()
	for _, obligation := range contract.Obligations {
		if obligation.ID != "TOOL-DEBUGGER" {
			continue
		}
		if obligation.Evidence.Source != SourceDebug || obligation.Evidence.Kind != EvidenceDebug ||
			obligation.Evidence.Counter != "paused_sessions" || obligation.Evidence.Target != 2 {
			t.Fatalf("debug evidence = %#v", obligation.Evidence)
		}
		return
	}
	t.Fatal("TOOL-DEBUGGER obligation missing")
}
