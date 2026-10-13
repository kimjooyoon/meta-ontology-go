package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func TestPackageActivityOriginPresence6971(t *testing.T) {
	const source = `package delivery6971
namespace delivery6971
entity Integer id "urn:gooo:type:integer"
activity Quote(Integer) -> Integer id "urn:delivery6971:quote" computes "if input > 103 { return input - 29 }; return input + 17"
`
	program, err := workspaceexecution.Prepare(packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "delivery6971", Activity: "Quote"},
		Packages: []packageruntime.PackageSpec{{Path: "delivery6971", Name: "delivery6971",
			Sources: []packageruntime.Source{{Filename: "배송6971.gooo", Content: source}}}}})
	if err != nil || len(program.ActivityOrigins) != 1 || len(program.ConditionOrigins) != 1 {
		t.Fatal("fresh declarations and conditions", err, program)
	}
	current, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(current)
	for _, form := range []struct {
		name       string
		omit       []string
		key, value string
		want       bool
	}{
		{name: "recorded", want: true},
		{name: "before-annotations", omit: []string{"activity_origins", "condition_origins"}, want: true},
		{name: "before-activity-origins", omit: []string{"activity_origins"}, want: true},
		{name: "activity-only", omit: []string{"condition_origins"}, want: true},
		{name: "explicit-null", key: "activity_origins", value: "null"},
		{name: "explicit-empty", key: "activity_origins", value: "[]"},
		{name: "different-declaration", key: "activity_origins", value: `[{"activity_id":"urn:other6971"}]`},
		{name: "different-source", omit: []string{"activity_origins", "condition_origins"}, key: "lowered_gooo_source", value: `"changed"`},
		{name: "different-identity", omit: []string{"activity_origins"}, key: "entry", value: `{"activity":"Other"}`},
		{name: "unknown-claim", key: "future_origins", value: "[]"},
	} {
		t.Run(form.name, func(t *testing.T) {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(current, &fields); err != nil {
				t.Fatal(err)
			}
			for _, key := range form.omit {
				delete(fields, key)
			}
			if form.key != "" {
				fields[form.key] = json.RawMessage(form.value)
			}
			recorded, err := json.Marshal(fields)
			if err != nil || samePackageProgramSource(recorded, current) != form.want {
				t.Fatal("program annotation presence", form.name, err)
			}
			receipt, err := json.Marshal(map[string]any{"result": map[string]any{"program": fields}})
			if err != nil || computationPackageReceiptProgramEqual(receipt, program) != form.want {
				t.Fatal("receipt annotation presence", form.name, err)
			}
			if len(form.omit) > 0 && form.want && samePackageProgramSource(current, recorded) {
				t.Fatal("recorded origin claim silently discarded")
			}
		})
	}
	if !bytes.Equal(before, current) {
		t.Fatal("comparison changed prepared source")
	}
	t.Log("6971: ten fresh annotation forms; absent fields retained, recorded declaration claims and source identity compared; preparation only; historical source/native/model executions0")
}
