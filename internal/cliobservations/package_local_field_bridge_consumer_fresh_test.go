package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// The same twelve source identities are checked in four three-input commands.
// This test is exercised by the successor's own original CI, never replayed
// locally after its historical observation.
func TestGoooFieldLocalBridge15317(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	example := filepath.Join(root, "examples/local-field-bridge")
	file, err := os.Open(filepath.Join(example, "recorded-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var recorded bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &recorded); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(t.TempDir(), "saved15317.json")
	if err := os.WriteFile(saved, raw, 0600); err != nil {
		t.Fatal(err)
	}
	offset, compared := 0, 0
	for c, trace := range recorded.Traces {
		var selected []int
		for i, use := range trace.FieldUses {
			site := recorded.FieldUseSites[use.SiteIndex]
			binding := recorded.FieldBindings[site.Binding]
			if binding.Name == "forwarded" && binding.ActivityID == "urn:gooo:local-bridge:build" {
				selected = append(selected, i)
			}
		}
		if len(selected) != 3 || selected[1] != selected[0]+1 || selected[2] != selected[1]+1 {
			t.Fatal("three consecutive final field identities", selected)
		}
		var out, diagnostic bytes.Buffer
		args := []string{"package", "execute", "--json", "--origin-receipt", saved, "--origin-shape", "field-local-columns", "--origin-offset", strconv.Itoa(offset + selected[0]), "--origin-limit", "3", filepath.Join(example, "trace.workspace.json")}
		if code := run(args, &out, &diagnostic); code != exitOK {
			t.Fatal("Gooo mixed-journal page", c, code, diagnostic.String(), out.String())
		}
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
			t.Fatal(err, receipt.Error)
		}
		origin := receipt.OriginInput
		if receipt.Decision != "OBSERVED" || receipt.Result.Runtime.ModelCalls != 0 || receipt.Result.Runtime.FiniteTotal != 0 || origin.InputShape != "gooo/field-local-lineage-columns/v1" || origin.ConsumerRows != 3 || origin.ReturnedRows != 3 || origin.ProducerExecutions != 0 || origin.ProducerModelCalls != 0 || origin.BridgeObservation == nil || len(receipt.Result.Runtime.Traces) != 3 {
			t.Fatal("record-only page scope", origin)
		}
		for j, index := range selected {
			site := recorded.FieldUseSites[trace.FieldUses[index].SiteIndex]
			binding := recorded.FieldBindings[site.Binding]
			checkBridgeAnswer15317(t, receipt.Result.Runtime.Traces[j], c, binding.Record.Fields[site.Field].ID)
			compared++
		}
		if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
			saveBridgePage15317(t, filepath.Join(dir, "consumer15317-page"+strconv.Itoa(c)+".json.gz"), out.Bytes())
		}
		offset += len(trace.FieldUses)
	}
	if compared != 12 {
		t.Fatal("expected twelve distinct saved field queries", compared)
	}
	t.Log("12 saved mixed-journal queries->four actual three-input Gooo commands12/12;all historical scalar-carrier assertions retained;producer/model0;input-only finite0")
}

func checkBridgeAnswer15317(t *testing.T, trace bodyexecution.CompositionTrace, c int, fieldID string) {
	t.Helper()
	var got struct {
		Status      string `json:"status"`
		Roots       string `json:"roots"`
		RootCount   int    `json:"root_count"`
		Carrier     int    `json:"carrier_writes"`
		FieldReads  int    `json:"field_read_edges"`
		LocalReads  int    `json:"local_read_edges"`
		Unresolved  int    `json:"unresolved_links"`
		Shared      int    `json:"shared_rhs_writes"`
		RecordingOK bool   `json:"recording_complete"`
	}
	if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &got) != nil {
		t.Fatal("mixed Gooo result shape")
	}
	input, minimum := "tag", 2
	if fieldID == "urn:gooo:local-bridge:score" {
		input = "left"
		if c%2 == 1 {
			input, minimum = "right", 1
		}
	} else if fieldID == "urn:gooo:local-bridge:live" {
		input = "enabled"
	}
	want := "urn:gooo:local-bridge:build|input|urn:gooo:local-bridge:input-" + input + ";"
	if got.Status != "SCOPED_FIELD_LOCAL_ORIGINS" || got.Roots != want || got.RootCount != 1 || got.Carrier < minimum || got.FieldReads < 1 || got.LocalReads < 1 || got.Unresolved != 0 || got.Shared != 0 || !got.RecordingOK {
		t.Fatal("Gooo-owned historical scalar-carrier origins", c, fieldID, got, want)
	}
}

func saveBridgePage15317(t *testing.T, path string, raw []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	z := gzip.NewWriter(f)
	if _, err := z.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
