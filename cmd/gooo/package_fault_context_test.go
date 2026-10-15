package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func savedPackageDivisionDiagnostic(t *testing.T) (bodyexecution.Composition, bodyexecution.CompositionRuntime) {
	t.Helper()
	raw, err := os.ReadFile("../../docs/observations/local-division-origins/first-observations.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Composition bodyexecution.Composition
		Runtime     bodyexecution.CompositionRuntime
	}
	entries := readNamedInputArchive(t, raw)
	if err := json.Unmarshal(entries["public-outputs/public-local-division.json"], &saved); err != nil {
		t.Fatal(err)
	}
	return saved.Composition, saved.Runtime
}

func TestPackageFaultContextUsesSavedGoooOrigins(t *testing.T) {
	composition, runtime := savedPackageDivisionDiagnostic(t)
	before, _ := json.Marshal(runtime)
	var stdout, stderr bytes.Buffer
	if code := writePackageEntryValues(&stdout, &stderr, "Divide", composition.Plan, runtime); code != exitFailure {
		t.Fatal("fault result", code)
	}
	if stdout.String() != "19\n-19\n" {
		t.Fatal("prior actual values changed", stdout.String())
	}
	for _, want := range []string{"input row 2", "ZERO_DIVISOR", `activity="urn:gooo:compound:divide"`, `operator="/" left=149 right=0`,
		`Go projection [169:185]: "amount /= 분모"`, `local="amount" assignment="/="`, `selected Gooo RHS [462:468]: "분모"`, runtime.SelectedSourceSHA256} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatal("missing retained context", want, stderr.String())
		}
	}
	after, _ := json.Marshal(runtime)
	if !bytes.Equal(before, after) {
		t.Fatal("diagnostic changed the saved execution")
	}
	if path := os.Getenv("GOOO_READ_OBSERVATION_DIR"); path != "" {
		if err := os.WriteFile(path+"/fault-diagnostic.txt", stderr.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPackageFaultContextKeepsUnmappedOriginsExplicit(t *testing.T) {
	for _, mode := range []string{"missing", "wrong-variable", "wrong-activity", "wrong-root", "wrong-operation", "wrong-projection", "wrong-range", "wrong-source", "unmapped", "duplicate", "binary", "field"} {
		_, runtime := savedPackageDivisionDiagnostic(t)
		fault := runtime.Traces[2].Deliveries[0].Fault.Site
		site := &runtime.LocalSites[1]
		switch mode {
		case "missing":
			site.Source.Value = nil
		case "wrong-variable":
			site.Name = "other"
		case "wrong-activity":
			site.ActivityID = "other"
		case "wrong-root":
			site.RootActivityID = "other"
		case "wrong-operation":
			site.Operation = "+="
		case "wrong-projection":
			site.ProjectionSHA256 = "other"
		case "wrong-range":
			site.End++
		case "wrong-source":
			site.Source.Value.SourceSHA256 = "other"
		case "unmapped":
			site.Source.Value.Status = "UNMAPPED"
		case "duplicate":
			runtime.LocalSites = append(runtime.LocalSites, *site)
		case "binary":
			fault.Assignment = ""
		case "field":
			fault.Variable = ""
		}
		if source := packageFaultLocalRHS(runtime, fault); source != nil {
			t.Fatal("invented Gooo source", mode, source)
		}
	}
}
