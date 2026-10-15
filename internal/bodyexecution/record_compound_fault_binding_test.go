package bodyexecution

import "testing"

func TestCompoundFieldFaultBindingRequiresCompleteOriginSet(t *testing.T) {
	for _, fixture := range []struct {
		name      string
		faults    []CompositionFaultSite
		wantError bool
	}{
		{"one", []CompositionFaultSite{{RootActivityID: "root", ActivityID: "helper", Assignment: "/="}}, false},
		{"missing", nil, true},
		{"wrong-source", []CompositionFaultSite{{RootActivityID: "root", ActivityID: "other", Assignment: "/="}}, true},
		{"duplicate", []CompositionFaultSite{{RootActivityID: "root", ActivityID: "helper", Assignment: "/="}, {RootActivityID: "root", ActivityID: "helper", Assignment: "/="}}, true},
	} {
		p := &compositionFieldProjection{inventory: compositionFieldInventory{bindings: []CompositionFieldBinding{{RootActivityID: "root", ActivityID: "helper"}}},
			sites: []CompositionFieldSite{{Operation: "/="}}, nodes: []compositionFieldNode{{kind: "compound_write"}}}
		if err := p.bindFieldFaults(fixture.faults); (err != nil) != fixture.wantError {
			t.Fatal(fixture.name, err)
		}
	}
}
