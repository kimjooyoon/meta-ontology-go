package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestMappedOutput13InterpreterFresh1905(t *testing.T) {
	source, err := os.ReadFile("../bodyexecution/testdata/text-bound1905/mapped-original.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte(`field canonical id "gooo://capability-probe/netip-mapped-output/result/v1/canonical" type string required one`)
	source = bytes.Replace(source, marker, append(append([]byte(nil), marker...), []byte(` max_utf8_bytes "1030"`)...), 1)
	generated, err := Generate("mapped.gooo", source, "MappedOutput")
	if err != nil {
		t.Fatal(err)
	}
	zone := strings.Repeat("z", 1007)
	input, _ := json.Marshal([]string{"::ffff:ffff:ffff%" + zone})
	expected, _ := json.Marshal(map[string]any{"parsed": true, "canonical": "::ffff:255.255.255.255%" + zone,
		"family": 6, "zone": zone, "is4_in6": true, "group0": 0, "group1": 0, "group2": 0,
		"group3": 0, "group4": 0, "group5": 65535, "group6": 65535, "group7": 65535})
	observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "MappedOutput", generated.Report.RecordTypes,
		[]assemblyspec.ValueCase{{Inputs: string(input), Expected: string(expected)}})
	if err != nil || len(observed) != 1 || !observed[0].Passed || len(observed[0].Fields) != 13 || !bytes.Equal(observed[0].Actual, expected) {
		t.Fatal("original typed thirteen-field interpretation differs", observed, err)
	}
}
