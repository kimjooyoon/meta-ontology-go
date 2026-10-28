package bodycodegen

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func compactInputPaths(r TypedValueOrigins, root uint16) []string {
	seen := make([]bool, len(r.Flow.Nodes)+1)
	seen[root] = true
	for _, node := range slices.Backward(r.Flow.Nodes) {
		if seen[node.ID] {
			seen[node.Parents[0]], seen[node.Parents[1]] = true, true
		}
	}
	var paths []string
	for _, input := range r.Inputs {
		if !seen[input.Node] {
			continue
		}
		// A container input is the origin prefix of its leaves, not another leaf.
		container := false
		for _, version := range r.Flow.RecordVersions {
			if version.Node == input.Node {
				container = true
			}
		}
		if container {
			continue
		}
		var ids []string
		for id := input.Node; id != 0; id = r.Flow.Nodes[id-1].Parents[0] {
			node := r.Flow.Nodes[id-1]
			if node.Kind != "input" {
				panic("input path contains a non-input node")
			}
			if node.FieldID != "" {
				ids = append(ids, node.FieldID)
			}
		}
		slices.Reverse(ids)
		paths = append(paths, input.Port+":"+strings.Join(ids, " > "))
	}
	slices.Sort(paths)
	return paths
}

func TestCompactActualOpenFGAInputAndOutputOrigins(t *testing.T) {
	source, err := os.ReadFile("../../examples/openfga-model-change/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := InspectTypedValueOrigins("comparison.gooo", source, "Compare")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	if r.Schema != "gooo/typed-value-origins/v2" || r.Flow.Schema != recordVersionFlowSchema ||
		r.SourceSHA256 != digest(source) || r.ActivityID != "gooo://authorization-delta/compare" ||
		len(r.Outputs) != 15 || len(r.Inputs) != 35 || len(r.Flow.Nodes) > 512 {
		t.Fatal("lost declared source extent")
	}
	before := "input:gooo://authorization-delta/pair/before > gooo://authorization-delta/observation/"
	after := "input:gooo://authorization-delta/pair/after > gooo://authorization-delta/observation/"
	for _, row := range []struct {
		field string
		want  []string
	}{
		{"before-input", []string{before + "input-origin"}},
		{"after-input", []string{after + "input-origin"}},
		{"case", []string{"input:gooo://authorization-delta/pair/case"}},
		{"origin-kind", []string{"input:gooo://authorization-delta/pair/origin-kind"}},
		{"before-error", []string{before + "error-code"}},
		{"after-error", []string{after + "error-code"}},
		{"input-match", []string{after + "input-origin", before + "input-origin"}},
		{"data-match", []string{after + "data", before + "data"}},
		{"execution-match", []string{after + "execution", before + "execution"}},
		{"model-match", []string{after + "model", after + "model-source", after + "requested-model", before + "model", before + "model-source", before + "requested-model", "input:gooo://authorization-delta/pair/after-model", "input:gooo://authorization-delta/pair/before-model"}},
	} {
		slices.Sort(row.want)
		t.Run(row.field, func(t *testing.T) {
			var root uint16
			for _, output := range r.Outputs {
				if output.FieldID == "gooo://authorization-delta/comparison/"+row.field {
					root = output.Root
				}
			}
			if root == 0 {
				t.Fatal("output not found")
			}
			if got := compactInputPaths(r, root); !reflect.DeepEqual(got, row.want) {
				t.Fatal("authored origin paths differ", got, row.want)
			}
		})
	}
	requireCompactVersionRoundTrip(t, r)
}

func requireCompactVersionRoundTrip(t *testing.T, r TypedValueOrigins) {
	t.Helper()
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TypedValueOrigins
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r, decoded) {
		t.Fatal("JSON round trip lost graph facts")
	}
	f := decoded.Flow
	if len(f.RecordLayouts) == 0 || len(f.RecordLayouts) > recordTypeLimit || len(f.RecordValues) > len(f.RecordVersions) || len(f.RecordVersions) > len(f.Nodes) {
		t.Fatal("invalid dictionary extent")
	}
	seen := map[uint16]bool{}
	uses := map[uint16]int{}
	expanded, stored := 0, 0
	for _, value := range f.RecordValues {
		if value.Layout == 0 || int(value.Layout) > len(f.RecordLayouts) {
			t.Fatal("dangling layout")
		}
		layout := f.RecordLayouts[value.Layout-1]
		if len(layout.Fields) != len(value.Fields) || len(value.Fields) > RecordFieldLimit {
			t.Fatal("field layout differs")
		}
		for _, root := range value.Fields {
			if root == 0 || int(root) > len(f.Nodes) {
				t.Fatal("dangling field root")
			}
		}
		stored += len(value.Fields)
	}
	for _, version := range f.RecordVersions {
		if version.Node == 0 || int(version.Node) > len(f.Nodes) || seen[version.Node] || version.Value == 0 || int(version.Value) > len(f.RecordValues) {
			t.Fatal("version identity lost")
		}
		seen[version.Node] = true
		uses[version.Value]++
		expanded += len(f.RecordValues[version.Value-1].Fields)
	}
	if len(uses) != len(f.RecordValues) || stored >= expanded {
		t.Fatal("actual consumer did not retain shared exact field-root vectors")
	}
	for _, n := range f.Nodes {
		if n.Parents[0] >= n.ID || n.Parents[1] >= n.ID || n.Guard >= n.ID || n.Condition >= n.ID {
			t.Fatal("dependency edges lost topological order")
		}
	}
	for _, n := range f.Nodes {
		if n.RecordVersion != 0 && (n.RecordVersion >= n.ID || !seen[n.RecordVersion]) {
			t.Fatal("dangling or future record version", n)
		}
	}
	for i, n := range f.Nodes {
		if n.Kind == "input" && n.CalleeID != "" {
			t.Fatal("shared input node was relabelled as a call")
		}
		if n.Span.View == "computes" && r.Sources[i].Status != "MAPPED" {
			t.Fatal("authored event lost its source span", n, r.Sources[i])
		}
	}
	t.Logf("nodes=%d versions=%d exact-value-vectors=%d stored-field-roots=%d expanded-field-roots=%d", len(f.Nodes), len(f.RecordVersions), len(f.RecordValues), stored, expanded)
}

func TestCompactVersionsKeepDifferentInputOriginsDistinct(t *testing.T) {
	r, err := InspectTypedValueOrigins("pairs.gooo", []byte(nominalOriginSource), "Both")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	values := map[string]uint16{}
	for _, node := range r.Flow.Nodes {
		if node.Kind != "input" || (node.FieldID != "origin://pair/before" && node.FieldID != "origin://pair/after") {
			continue
		}
		for _, version := range r.Flow.RecordVersions {
			if version.Node == node.ID {
				values[node.FieldID] = version.Value
			}
		}
	}
	if len(values) != 2 || values["origin://pair/before"] == values["origin://pair/after"] {
		t.Fatal("different original paths share a value identity", values)
	}
}

func TestCompactUnusedLayoutPreservesLegacyGraph(t *testing.T) {
	source := []byte(typedSource12973)
	before, err := InspectTypedValueOrigins("same.gooo", source, "Copied")
	if err != nil {
		t.Fatal(err)
	}
	added := append(append([]byte(nil), source...), []byte("\nentity Unused id \"origin://unused\" fields { field child id \"origin://unused/child\" type Pair required one }\n")...)
	after, err := InspectTypedValueOrigins("same.gooo", added, "Copied")
	if err != nil || after.Status != "RESOLVED" {
		t.Fatal(err, after.Reason)
	}
	if before.SourceSHA256 == after.SourceSHA256 || !reflect.DeepEqual(before.Flow, after.Flow) || before.Scope != after.Scope || len(after.Flow.RecordVersions) != 0 {
		t.Fatal("unused nominal layout changed legacy graph identity")
	}
}

func TestCompactRecordHelperKeepsCallsAndInputsSeparate(t *testing.T) {
	source := nominalOriginSource + "\nactivity Carry(pair Pair) -> Pair id \"origin://carry\" computes \"return pair\"\nactivity Twice(pair Pair) -> Integer id \"origin://twice\" computes \"let first=Carry(pair);let second=Carry(pair);return first.before.value+second.after.value\"\n"
	r, err := InspectTypedValueOrigins("helper.gooo", []byte(source), "Twice")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	want := []string{"input:origin://pair/after > origin://leaf/value", "input:origin://pair/before > origin://leaf/value"}
	if got := compactInputPaths(r, r.Outputs[0].Root); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	versions := map[uint16]uint16{}
	for _, v := range r.Flow.RecordVersions {
		versions[v.Node] = v.Value
	}
	var calls []uint16
	for _, n := range r.Flow.Nodes {
		if n.Kind == "input" && n.CalleeID != "" {
			t.Fatal("callee rewrote shared original input", n)
		}
		if n.Kind == "call" && n.CalleeID == "origin://carry" {
			calls = append(calls, n.ID)
		}
	}
	if len(calls) != 2 || calls[0] == calls[1] || versions[calls[0]] == 0 || versions[calls[0]] != versions[calls[1]] {
		t.Fatal("call instances or exact shared payload lost", calls, versions)
	}
	if _, err := TypedValueOriginInputs(r, "reader:Inputs", "", false); err == nil {
		t.Fatal("flat columns dropped record version facts")
	}
	if _, err := TypedSourceControlInputs(r, "reader:Inputs"); err == nil {
		t.Fatal("control columns dropped record version facts")
	}
}

func TestCompactStillEnforcesActualNodeBudget(t *testing.T) {
	source := nominalOriginSource + "\nactivity TooMany(pair Pair) -> Integer id \"origin://too-many\" computes `let n=0;" + strings.Repeat("n=n+1;", 200) + "return pair.before.value+n`\n"
	r, err := InspectTypedValueOrigins("budget.gooo", []byte(source), "TooMany")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_NODE_BOUND" || len(r.Outputs) != 0 || len(r.Flow.Nodes) != 0 || len(r.Flow.RecordVersions) != 0 {
		t.Fatal("budget failed to close whole result", err, r)
	}
}

func TestCompactRecordHistoryIsSeparateFromFieldDependencies(t *testing.T) {
	r, err := InspectTypedValueOrigins("preserved.gooo", []byte(nominalOriginSource), "Preserve")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	data := compactReachable(r, r.Outputs[0].Root, false, false)
	history := compactReachable(r, r.Outputs[0].Root, false, true)
	found := false
	for i, node := range r.Flow.Nodes {
		if node.Kind == "write" && strings.Contains(r.Sources[i].Expression, "moved.before") {
			if data[node.ID] {
				t.Fatal("unrelated field update became a value dependency")
			}
			found = found || history[node.ID]
		}
	}
	if !found {
		t.Fatal("record-version projection lost the actual update event")
	}
	r, err = InspectTypedValueOrigins("conditional.gooo", []byte(nominalOriginSource), "Choose")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	data = compactReachable(r, r.Outputs[0].Root, false, false)
	control := compactReachable(r, r.Outputs[0].Root, true, false)
	found = false
	for _, input := range r.Inputs {
		if input.Port == "input2" {
			if data[input.Node] || !control[input.Node] {
				t.Fatal("conditional value lost its separate control dependency")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("original condition input is missing")
	}
}

func compactReachable(r TypedValueOrigins, root uint16, control, versions bool) []bool {
	seen := make([]bool, len(r.Flow.Nodes)+1)
	seen[root] = true
	for _, node := range slices.Backward(r.Flow.Nodes) {
		if !seen[node.ID] {
			continue
		}
		seen[node.Parents[0]], seen[node.Parents[1]] = true, true
		if control {
			seen[node.Guard], seen[node.Condition] = true, true
		}
		if versions {
			seen[node.RecordVersion] = true
		}
	}
	return seen
}

func TestCompactNodeMetadataFitsExistingPadding(t *testing.T) {
	for _, row := range []struct {
		typeInfo reflect.Type
		field    string
	}{
		{reflect.TypeFor[RecordFlowNode](), "RecordVersion"},
		{reflect.TypeFor[recordFlowBuilder](), "compactRecords"},
	} {
		var baseline []reflect.StructField
		for i := range row.typeInfo.NumField() {
			field := row.typeInfo.Field(i)
			if field.Name != row.field {
				baseline = append(baseline, field)
			}
		}
		if row.typeInfo.Size() != reflect.StructOf(baseline).Size() {
			t.Fatal("metadata added padding to the existing node or mode layout", row.typeInfo, row.field)
		}
	}
	t.Logf("record-flow node bytes=%d", reflect.TypeFor[RecordFlowNode]().Size())
}

func TestCompactVersionsDeclineLegacyModelFeatures(t *testing.T) {
	r, err := InspectTypedValueOrigins("pair.gooo", []byte(nominalOriginSource), "Both")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	for _, encoded := range []*RecordOrdinalContext{
		recordOriginContext(nil, r.Flow),
		recordGraphContext(recordAssemblyPlan{}, r.Flow),
	} {
		if encoded.Status != "DECLINED_TO_DETERMINISTIC" ||
			encoded.Reason != "RECORD_VERSION_FEATURE_CONTRACT_UNAVAILABLE" ||
			encoded.ValueFlowSHA256 == "" || encoded.Text != "" || len(encoded.Parts) != 0 {
			t.Fatal("legacy model contract lost record-version facts", encoded)
		}
	}
}
