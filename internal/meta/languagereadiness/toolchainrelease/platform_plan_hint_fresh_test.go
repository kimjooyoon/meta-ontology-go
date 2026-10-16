package toolchainrelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPlanInspectionAdvisoryBoundaryFresh1882(t *testing.T) {
	example := planInputExamples()[0]
	source, err := os.ReadFile(filepath.Join("../../../..", example.source))
	if err != nil {
		t.Fatal(err)
	}
	raw := planInputsFixture(t, "record-korean-plan")
	for _, mode := range []string{"absent", "large_hint", "null", "text", "wrong_field_type", "future_schema",
		"changed_source", "missing_counter", "nonzero_counter", "changed_plan"} {
		t.Run(mode, func(t *testing.T) {
			var view map[string]json.RawMessage
			if err := json.Unmarshal(raw, &view); err != nil {
				t.Fatal(err)
			}
			view["transport_capacity"] = json.RawMessage(`{"input_json_bytes":99999999}`)
			invalid := false
			switch mode {
			case "absent":
				delete(view, "transport_capacity")
			case "null":
				view["transport_capacity"] = json.RawMessage(`null`)
			case "text":
				view["transport_capacity"] = json.RawMessage(`"{unreadable advisory content"`)
			case "wrong_field_type":
				view["transport_capacity"] = json.RawMessage(`{"input_json_bytes":"not an integer"}`)
			case "future_schema":
				view["transport_capacity"] = json.RawMessage(`{"schema":"future/hint/v99"}`)
			case "changed_source":
				view["source_sha256"], invalid = json.RawMessage(`"sha256:changed"`), true
			case "missing_counter":
				delete(view, "model_calls")
				invalid = true
			case "nonzero_counter":
				view["native_executions"], invalid = json.RawMessage(`1`), true
			case "changed_plan":
				var plan map[string]json.RawMessage
				if err := json.Unmarshal(view["plan"], &plan); err != nil {
					t.Fatal(err)
				}
				plan["entry_activity"] = json.RawMessage(`"Other"`)
				view["plan"], err = json.Marshal(plan)
				if err != nil {
					t.Fatal(err)
				}
				invalid = true
			}
			candidate, err := json.Marshal(view)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validatePlanInputInspection(candidate, source, example)
			if (err != nil) != invalid {
				t.Fatal("hint crossed the structural plan/counter boundary", mode, err)
			}
		})
	}
}
