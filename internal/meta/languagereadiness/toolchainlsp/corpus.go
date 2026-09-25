package toolchainlsp

import "fmt"

var caseContract = []Case{
	{ID: "server-lifecycle", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "initialize-capabilities", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "diagnostics-clean", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "hover", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "completion", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "definition", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "document-symbol", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "workspace-symbol", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "references", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "semantic-tokens", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "diagnostics-malformed", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "unsupported-method", Group: "PROTOCOL", Expected: "FAIL_CLOSED"},
	{ID: "close-clears-diagnostics", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "shutdown-exit", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "transcript-replay", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "utf16-roundtrip", Group: "PROTOCOL", Expected: "PASS"},
	{ID: "coupling-pass", Group: "COUPLING", Expected: "PASS"},
	{ID: "coupling-upstream-unknown", Group: "COUPLING", Expected: "LOWER_RESOLUTION"},
	{ID: "coupling-upstream-fail-closed", Group: "COUPLING", Expected: "FAIL_CLOSED"},
	{ID: "coupling-stale-snapshot", Group: "COUPLING", Expected: "LOWER_RESOLUTION"},
	{ID: "coupling-cancelled", Group: "COUPLING", Expected: "LOWER_RESOLUTION"},
	{ID: "coupling-input-immutability", Group: "COUPLING", Expected: "PASS"},
}

func CanonicalCorpus() Corpus {
	return Corpus{Schema: CorpusSchema, Cases: append([]Case(nil), caseContract...)}
}

func ValidateCorpus(corpus Corpus) error {
	if corpus.Schema != CorpusSchema || len(corpus.Cases) != len(caseContract) {
		return fmt.Errorf("corpus schema/count mismatch")
	}
	for index, expected := range caseContract {
		if corpus.Cases[index] != expected {
			return fmt.Errorf("corpus case %d drift", index)
		}
	}
	return nil
}
