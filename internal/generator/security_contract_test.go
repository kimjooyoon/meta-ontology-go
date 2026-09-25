package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestMarkerSecurityContractEvidence(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	source := string(result.Source)
	cases := []struct {
		id   string
		data []byte
	}{
		{"GEN-MARK-001 unknown generated attribute", []byte(strings.Replace(source, `kind="entity"`, `kind="entity" extra="x"`, 1))},
		{"GEN-MARK-002 unknown slot attribute", []byte(strings.Replace(source, `//gooo:slot:start id="gooo://slot/compile-implementation"`, `//gooo:slot:start id="gooo://slot/compile-implementation" extra="x"`, 1))},
		{"GEN-MARK-003 missing generated ID", []byte(strings.Replace(source, `//gooo:generated:end id="gooo://activity/compile" kind="activity"`, `//gooo:generated:end kind="activity"`, 1))},
		{"GEN-MARK-004 missing generated kind", []byte(strings.Replace(source, `//gooo:generated:start id="gooo://activity/compile" kind="activity"`, `//gooo:generated:start id="gooo://activity/compile"`, 1))},
		{"GEN-MARK-005 mismatched generated kind", []byte(strings.Replace(source, `//gooo:generated:end id="gooo://activity/compile" kind="activity"`, `//gooo:generated:end id="gooo://activity/compile" kind="entity"`, 1))},
		{"GEN-MARK-006 nested region", append(append([]byte(nil), result.Source...), []byte("\n//gooo:generated:start id=\"nested-a\" kind=\"entity\"\n//gooo:generated:start id=\"nested-b\" kind=\"entity\"\n")...)},
		{"GEN-MARK-007 orphan slot", append(append([]byte(nil), result.Source...), []byte("\n//gooo:slot:start id=\"orphan\"\n")...)},
		{"GEN-MARK-008 duplicate region", append(append([]byte(nil), result.Source...), testGeneratedBlock(t, result.Source, "gooo://entity/source")...)},
		{"GEN-MARK-009 duplicate slot", append(append([]byte(nil), result.Source...), []byte("\n//gooo:generated:start id=\"extra\" kind=\"activity\"\n//gooo:slot:start id=\"gooo://slot/compile-implementation\"\n//gooo:slot:end id=\"gooo://slot/compile-implementation\"\n//gooo:generated:end id=\"extra\" kind=\"activity\"\n")...)},
		{"GEN-MARK-010 region-slot collision", []byte(strings.Replace(strings.Replace(source, `//gooo:slot:start id="gooo://slot/compile-implementation"`, `//gooo:slot:start id="gooo://entity/source"`, 1), `//gooo:slot:end id="gooo://slot/compile-implementation"`, `//gooo:slot:end id="gooo://entity/source"`, 1))},
		{"GEN-MARK-011 invalid line boundary", []byte(strings.Replace(source, `//gooo:generated:start id="gooo://entity/source"`, `//gooo:generated:startX id="gooo://entity/source"`, 1))},
	}
	for _, testCase := range cases {
		t.Run(testCase.id, func(t *testing.T) {
			assertRejectedPrevious(t, testCase.data)
		})
	}
}

func TestNoWriteContractEvidence(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := append(append([]byte(nil), result.Source...), []byte("\n//gooo:slot:start id=\"orphan\"\n")...)
	// GEN-NOWRITE-001: malformed marker rejection preserves previous bytes.
	assertRejectedPrevious(t, previous)

	// GEN-NOWRITE-002: IR rejection preserves the caller-owned IR.
	ir := acceptanceFixture()
	ir.Activities[0].Slots[0].ID = ir.Entities[0].ID
	before := copyIR(ir)
	if _, err := Generate(ir, nil); err == nil {
		t.Fatal("invalid IR was accepted")
	}
	if !reflect.DeepEqual(ir, before) {
		t.Fatal("rejected generation changed caller-owned IR")
	}

	// GEN-NOWRITE-003: legacy rejection returns no replacement output.
	malformed := BeginMarker + "old\nbody\n"
	if output, err := MergeGenerated(malformed, malformed); err == nil || output != "" {
		t.Fatalf("legacy rejection produced output=%q err=%v", output, err)
	}
}

func assertRejectedPrevious(t *testing.T, previous []byte) {
	t.Helper()
	original := append([]byte(nil), previous...)
	if _, err := Generate(acceptanceFixture(), previous); err == nil {
		t.Fatal("corrupted previous source was accepted")
	}
	if !bytes.Equal(previous, original) {
		t.Fatal("rejected generation changed caller-owned previous bytes")
	}
}
