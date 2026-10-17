package cliobservations

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type mappedInputFixture1907 struct {
	source, manifest, inputs []byte
	workspace, inputPath     string
}

func TestBoundedMappedOriginalInputOnlyNativeFresh1907(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	fixture := mappedOriginalInputs1907(t)
	var output, diagnostic bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--inputs", fixture.inputPath,
		fixture.workspace}, &output, &diagnostic)
	retainTextBoundScope1907(t, "text-bound-input-only", "mapped13", map[string]any{
		"source": string(fixture.source), "manifest_raw": string(fixture.manifest),
		"inputs_raw": string(fixture.inputs), "exit_code": code,
		"stdout": output.String(), "stderr": diagnostic.String(),
	})
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil {
		t.Fatal("input-only receipt is not JSON", code, err, output.String(), diagnostic.String())
	}
	if code != exitOK || receipt.Result == nil || receipt.Error != "" || receipt.Decision != "OBSERVED" {
		t.Fatal("declared bounded output was not observed", code, receipt.Error, output.String())
	}
	digest := sha256.Sum256(fixture.inputs)
	if receipt.InputsDigest != "sha256:"+hex.EncodeToString(digest[:]) || receipt.CasesDigest != "" {
		t.Fatal("input-only origin was relabeled as scored cases", receipt.InputsDigest, receipt.CasesDigest)
	}
	assertMappedInputRuntime1907(t, receipt.Result.Runtime)
}

func mappedOriginalInputs1907(t *testing.T) mappedInputFixture1907 {
	t.Helper()
	dir := t.TempDir()
	fixture := mappedInputFixture1907{source: boundedMappedOriginalSource1907(t),
		workspace: filepath.Join(dir, "mapped-output.workspace.json"),
		inputPath: filepath.Join(dir, "00-port.inputs.json")}
	root := "../bodyexecution/testdata/text-bound1905/"
	var err error
	fixture.manifest, err = os.ReadFile(root + "mapped-original.workspace.json.fixture")
	if err != nil {
		t.Fatal(err)
	}
	fixture.inputs, err = os.ReadFile(root + "mapped-original.inputs.json.fixture")
	if err != nil {
		t.Fatal(err)
	}
	requireMappedDigest1907(t, fixture.manifest, "e4ad88f9c81b24f593aacac0d3263523924556e5e53b8bd385a2d31eafb6a7e3")
	requireMappedDigest1907(t, fixture.inputs, "cd9a55beec3ceaab3d119a5da903b4165375016d8f8795d4c5b7eff326c21481")
	if len(fixture.inputs) != 1130 {
		t.Fatal("original input document bytes changed")
	}
	for name, raw := range map[string][]byte{"mapped-output.gooo": fixture.source,
		"mapped-output.workspace.json": fixture.manifest, "00-port.inputs.json": fixture.inputs} {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return fixture
}

func boundedMappedOriginalSource1907(t *testing.T) []byte {
	t.Helper()
	original, err := os.ReadFile("../bodyexecution/testdata/text-bound1905/mapped-original.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	requireMappedDigest1907(t, original, "1c23bfc07dfb68b5d03b762a572f790dcb43f2546b312dd8d02ae3ab909bdab5")
	marker := []byte(`field canonical id "gooo://capability-probe/netip-mapped-output/result/v1/canonical" type string required one`)
	if bytes.Count(original, marker) != 1 {
		t.Fatal("original canonical declaration is ambiguous")
	}
	return bytes.Replace(original, marker, append(append([]byte(nil), marker...), []byte(` max_utf8_bytes "1030"`)...), 1)
}

func requireMappedDigest1907(t *testing.T, raw []byte, wanted string) {
	t.Helper()
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != wanted {
		t.Fatal("original issue1620 bytes changed")
	}
}

func assertMappedInputRuntime1907(t *testing.T, runtime bodyexecution.CompositionRuntime) {
	t.Helper()
	if runtime.Stage != "COMPLETE" || !runtime.ProjectionReplayed || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 ||
		runtime.FiniteTotal != 0 || runtime.FinitePassed != 0 || len(runtime.Traces) != 1 || len(runtime.Runs) != 2 {
		t.Fatal("input-only observation was scored or incompletely executed", runtime)
	}
	if !textBoundProducerMatches1906(runtime.ProducerSourceSHA, commandSourceIdentity, os.Getenv("GOOO_CI_SOURCE_HEAD")) {
		t.Fatal("input-only producer differs from actual compiler build", runtime.ProducerSourceSHA)
	}
	if !runtime.Build.Started || !runtime.Build.Completed || runtime.Build.ExitCode == nil ||
		*runtime.Build.ExitCode != 0 || runtime.Build.TimedOut || runtime.Build.Canceled || runtime.Build.OutputTruncated {
		t.Fatal("input-only native build incomplete", runtime.Build)
	}
	if runtime.InputSeparation.Status != "UNKNOWN" || runtime.InputSeparation.DisjointCasesPassed != 0 {
		t.Fatal("input-only observation acquired evaluation credit", runtime.InputSeparation)
	}
	for _, process := range runtime.Runs {
		if !process.Started || !process.Completed || process.ExitCode == nil || *process.ExitCode != 0 ||
			process.TimedOut || process.Canceled || process.OutputTruncated {
			t.Fatal("input-only native execution incomplete", process)
		}
	}
	if runtime.Runs[0].StdoutSHA256 != runtime.Runs[1].StdoutSHA256 {
		t.Fatal("native observations differ")
	}
	if len(runtime.Traces[0].Deliveries) != 1 {
		t.Fatal("original mapped delivery missing")
	}
	delivery := runtime.Traces[0].Deliveries[0]
	if len(delivery.Expected) != 0 || delivery.Passed != nil || delivery.Fault != nil || len(delivery.BlockedBy) != 0 {
		t.Fatal("input-only observation fabricated an oracle or retained a fault", delivery)
	}
	assertMappedInputValues1907(t, delivery)
}

func assertMappedInputValues1907(t *testing.T, delivery bodyexecution.CompositionDelivery) {
	t.Helper()
	zone := strings.Repeat("z", 1007)
	if len(delivery.Inputs) != 0 || len(delivery.Input) == 0 {
		t.Fatal("original unary input shape changed")
	}
	var input string
	if json.Unmarshal(delivery.Input, &input) != nil || input != "::ffff:ffff:ffff%"+zone || len(input) != 1024 {
		t.Fatal("original input value changed", delivery.Input)
	}
	names := []string{"parsed", "canonical", "family", "zone", "is4_in6", "group0", "group1",
		"group2", "group3", "group4", "group5", "group6", "group7"}
	wanted := map[string]any{"parsed": true, "canonical": "::ffff:255.255.255.255%" + zone,
		"family": 6, "zone": zone, "is4_in6": true, "group0": 0, "group1": 0, "group2": 0,
		"group3": 0, "group4": 0, "group5": 65535, "group6": 65535, "group7": 65535}
	var actual map[string]json.RawMessage
	if json.Unmarshal(delivery.Actual, &actual) != nil || len(actual) != 13 || len(delivery.ActualFields) != 13 {
		t.Fatal("original thirteen typed fields missing", delivery)
	}
	for i, name := range names {
		field := delivery.ActualFields[i]
		value, _ := json.Marshal(wanted[name])
		if field.Name != name || field.ID != "gooo://capability-probe/netip-mapped-output/result/v1/"+name ||
			!bytes.Equal(field.Value, value) || !bytes.Equal(actual[name], value) {
			t.Fatal("original field identity/order/type/value changed", field)
		}
	}
}
