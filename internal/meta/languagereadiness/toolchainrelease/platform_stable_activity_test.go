package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stableActivityFixture(entry string) []byte {
	const identity = "urn:gooo:example:total"
	rows := [4][3]string{{"2", "3", "5"}, {"-7", "2", "0"}, {"0", "0", "0"}, {"9007199254740993", "7", "9007199254741000"}}
	traces := make([]string, 0, len(rows))
	for i, row := range rows {
		traces = append(traces, fmt.Sprintf(`{"case_index":%d,"deliveries":[{"activity_id":%q,"inputs":[{"port":"input0","entity_id":"urn:gooo:type:integer","value":%s},{"port":"input1","entity_id":"urn:gooo:type:integer","value":%s}],"actual":%s,"expected":%s,"passed":true}]}`, i, identity, row[0], row[1], row[2], row[2]))
	}
	source, _ := json.Marshal("//gooo:generated:start id=\"" + identity + "\" kind=\"activity\"\n//gooo:generated:end id=\"" + identity + "\" kind=\"activity\"\n")
	return []byte(fmt.Sprintf(`{"composition":{"source":%s,"plan":{"entry_activity":%q,"activities":[{"name":%q,"id":%q,"inputs":[{"port":"input0","entity_id":"urn:gooo:type:integer","from":-1},{"port":"input1","entity_id":"urn:gooo:type:integer","from":-1}]}]}},"runtime":{"traces":[%s]}}`, source, entry, entry, identity, strings.Join(traces, ",")))
}

func stableActivityMutations() map[string]nativeSmokeEdit {
	return map[string]nativeSmokeEdit{
		"entry":         {[]any{"composition", "plan", "entry_activity"}, "Other"},
		"name":          {[]any{"composition", "plan", "activities", 0, "name"}, "Other"},
		"plan ID":       {[]any{"composition", "plan", "activities", 0, "id"}, "stable://activity/total"},
		"raw ID":        {[]any{"composition", "plan", "activities", 0, "id"}, "URN:gooo:example:total"},
		"input order":   {[]any{"composition", "plan", "activities", 0, "inputs", 0, "port"}, "input1"},
		"marker":        {[]any{"composition", "source"}, "package main\n"},
		"traces":        {[]any{"runtime", "traces"}, nil},
		"trace order":   {[]any{"runtime", "traces", 0, "case_index"}, json.Number("3")},
		"delivery ID":   {[]any{"runtime", "traces", 0, "deliveries", 0, "activity_id"}, "other://activity/total"},
		"port":          {[]any{"runtime", "traces", 0, "deliveries", 0, "inputs", 0, "port"}, "input1"},
		"input type":    {[]any{"runtime", "traces", 0, "deliveries", 0, "inputs", 0, "entity_id"}, "urn:gooo:type:boolean"},
		"rounded input": {[]any{"runtime", "traces", 3, "deliveries", 0, "inputs", 0, "value"}, json.Number("9007199254740992")},
		"result":        {[]any{"runtime", "traces", 3, "deliveries", 0, "actual"}, json.Number("9007199254741001")},
		"expected":      {[]any{"runtime", "traces", 0, "deliveries", 0, "expected"}, json.Number("6")},
		"pass omitted":  {[]any{"runtime", "traces", 0, "deliveries", 0, "passed"}, nil},
		"pass":          {[]any{"runtime", "traces", 0, "deliveries", 0, "passed"}, false},
		"fault":         {[]any{"runtime", "traces", 0, "deliveries", 0, "fault"}, map[string]any{}},
	}
}

func TestStableActivitySmokeRequiresIdentityAndExactNativeValues(t *testing.T) {
	for _, entry := range []string{"Total", "합계"} {
		raw := stableActivityFixture(entry)
		if err := validateStableActivitySmoke(raw, entry); err != nil {
			t.Fatal(entry, err)
		}
		for name, edit := range stableActivityMutations() {
			t.Run(entry+"/"+name, func(t *testing.T) {
				if err := validateStableActivitySmoke(changedNativeSmoke(t, raw, edit), entry); err == nil {
					t.Fatal("changed identity or native observation accepted")
				}
			})
		}
	}
	for _, raw := range [][]byte{nil, []byte(`{}`), append(stableActivityFixture("Total"), []byte(`{}`)...)} {
		if err := validateStableActivitySmoke(raw, "Total"); err == nil {
			t.Fatal("incomplete native observation accepted")
		}
	}
}

func TestNativeStableActivityReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if err := smokeStableActivity(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
	for _, example := range stableActivityExamples() {
		for _, mode := range []string{"construct", "replay"} {
			raw, err := os.ReadFile(filepath.Join(input.OutputDir, input.Target.ID+"-"+example.name+"-"+mode+".json"))
			if err != nil {
				t.Fatal(err)
			}
			for name, edit := range stableActivityMutations() {
				t.Run(example.name+"/"+mode+"/"+name, func(t *testing.T) {
					if err := validateStableActivitySmoke(changedNativeSmoke(t, raw, edit), example.entry); err == nil {
						t.Fatal("changed real native observation accepted")
					}
				})
			}
		}
	}
}
