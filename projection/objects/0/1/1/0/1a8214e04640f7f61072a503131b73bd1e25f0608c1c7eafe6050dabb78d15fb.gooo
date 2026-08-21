package roundtrip

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestMarkerRejectsSlotAndRegionIdentityCollisions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "duplicate slot across regions",
			source: `package billinggen
//gooo:generated:start id="billing://entity/first" kind="entity"
//gooo:slot:start id="billing://slot/shared"
func First() {}
//gooo:slot:end id="billing://slot/shared"
//gooo:generated:end id="billing://entity/first" kind="entity"
//gooo:generated:start id="billing://entity/second" kind="entity"
//gooo:slot:start id="billing://slot/shared"
func Second() {}
//gooo:slot:end id="billing://slot/shared"
//gooo:generated:end id="billing://entity/second" kind="entity"
`,
		},
		{
			name: "slot nested in region",
			source: `package billinggen
//gooo:generated:start id="billing://shared" kind="entity"
//gooo:slot:start id="billing://shared"
func Shared() {}
//gooo:slot:end id="billing://shared"
//gooo:generated:end id="billing://shared" kind="entity"
`,
		},
		{
			name: "region after slot",
			source: `package billinggen
//gooo:generated:start id="billing://entity/first" kind="entity"
//gooo:slot:start id="billing://shared"
func First() {}
//gooo:slot:end id="billing://shared"
//gooo:generated:end id="billing://entity/first" kind="entity"
//gooo:generated:start id="billing://shared" kind="entity"
type Shared struct{}
//gooo:generated:end id="billing://shared" kind="entity"
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := CheckLocality(LocalityInput{Before: []byte(test.source), After: []byte(test.source)})
			if report.OK() || report.Violations[0].Rule != RuleMarker {
				t.Fatalf("identity collision was not reported: %s", report.Error())
			}
		})
	}
}
func TestMalformedSemanticSnapshotsAreReported(t *testing.T) {
	report := CheckDSLToIR(semantic.IR{}, MinimalFixture().IR)
	if report.OK() || report.Violations[0].Rule != RuleSnapshot {
		t.Fatalf("malformed semantic snapshot was not reported: %s", report.Error())
	}
}
