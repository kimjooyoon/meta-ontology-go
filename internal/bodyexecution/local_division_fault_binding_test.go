package bodyexecution

import "testing"

func TestLocalDivisionFaultBindingRequiresExactVariable(t *testing.T) {
	for _, row := range []struct {
		name, variable string
		count          int
		wantError      bool
	}{
		{"valid", "amount", 1, false}, {"wrong-variable", "other", 1, true},
		{"missing", "amount", 0, true}, {"extra", "amount", 2, true}, {"field-only", "", 1, true},
	} {
		p := readProjection{localSites: []CompositionLocalSite{{RootActivityID: "root", ActivityID: "helper", Name: "amount", Operation: "/="}},
			nodes: []readProjectionNode{{localKind: "compound_write"}}}
		faults := make([]CompositionFaultSite, row.count)
		for i := range faults {
			faults[i] = CompositionFaultSite{RootActivityID: "root", ActivityID: "helper", Assignment: "/=", Variable: row.variable}
		}
		if err := p.bindLocalFaults(faults); (err != nil) != row.wantError {
			t.Fatal(row.name, err)
		}
	}
}
