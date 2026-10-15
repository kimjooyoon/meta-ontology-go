package bodycodegen

import "testing"

func TestNamedColumnsRejectSourceDisagreement(t *testing.T) {
	mutations := map[string]func(*TypedValueOrigins){
		"source digest": func(r *TypedValueOrigins) { r.SourceSHA256 = "other" },
		"name text":     func(r *TypedValueOrigins) { r.Inputs[0].Declaration.Name = "other" },
		"name range":    func(r *TypedValueOrigins) { r.Inputs[0].Declaration.NameEnd++ },
		"unnamed span":  func(r *TypedValueOrigins) { r.Inputs[1].Declaration.NameStart = 1 },
		"type text": func(r *TypedValueOrigins) {
			r.Inputs[0].Declaration.TypeName, r.Inputs[0].Declaration.Expression = "Other", "Other"
		},
		"type range":            func(r *TypedValueOrigins) { r.Inputs[0].Declaration.End = len(namedColumnsSource) + 1 },
		"missing type identity": func(r *TypedValueOrigins) { r.Inputs[0].Declaration.TypeID = "" },
		"noncanonical port": func(r *TypedValueOrigins) {
			r.Inputs[0].Declaration.Port, r.Inputs[0].Port = "기준", "기준"
		},
		"missing position": func(r *TypedValueOrigins) { r.Inputs = r.Inputs[1:] },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			r := namedColumnsReport(t, "Delta")
			change(&r)
			if raw, err := TypedNamedInputDeclarationInputs(r, []byte(namedColumnsSource), "source/control:DescribeInputs"); err == nil || len(raw) != 0 {
				t.Fatal("inconsistent source facts exported", err)
			}
		})
	}
	r := namedColumnsReport(t, "Read")
	r.Inputs[1].Declaration.Name = "different"
	if _, err := TypedNamedInputDeclarationInputs(r, []byte(namedColumnsSource), "source/control:DescribeInputs"); err == nil {
		t.Fatal("record fields disagree about one signature position")
	}
	if _, err := TypedNamedInputDeclarationInputs(namedColumnsReport(t, "Delta"), []byte(namedColumnsSource+"\n"), "source/control:DescribeInputs"); err == nil {
		t.Fatal("stale source accepted")
	}
}
