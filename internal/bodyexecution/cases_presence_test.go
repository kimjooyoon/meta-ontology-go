package bodyexecution

import "testing"

func TestRuntimeCasesRequireExplicitValues(t *testing.T) {
	for _, c := range []string{`{}`, `{"input":0}`, `{"expected":0}`, `{"input":null,"expected":0}`, `{"input":0,"expected":null}`} {
		if _, err := DecodeCases([]byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[` + c + `]}`)); err == nil {
			t.Fatal("missing/null value silently became zero", c)
		}
	}
	cases, err := DecodeCases([]byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0,"expected":0}]}`))
	if err != nil || len(cases) != 1 || cases[0].Input != 0 || cases[0].Expected != 0 {
		t.Fatal("explicit zero rejected", err)
	}
}
