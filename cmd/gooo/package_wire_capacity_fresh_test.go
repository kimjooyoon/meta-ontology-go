package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPackageWireCapacityFresh1880(t *testing.T) {
	source, err := os.ReadFile("../../internal/bodyexecution/testdata/wire-capacity.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{
		"wire/main.gooo.fixture":   source,
		"wire/gooo.workspace.json": []byte(`{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"example/wire","activity":"Second"},"packages":[{"path":"example/wire","name":"wire","imports":[],"sources":["main.gooo.fixture"]}]}`),
	}
	for _, mode := range []string{"inspect", "read", "over"} {
		t.Run(mode, func(t *testing.T) {
			if mode == "inspect" {
				var out, diagnostic bytes.Buffer
				if code := runPackageInspect([]string{"--json", "wire/gooo.workspace.json"}, reader, &out, &diagnostic); code != exitOK {
					t.Fatal(code, diagnostic.String(), out.String())
				}
				var r workspaceInspectionReceipt
				if err := json.Unmarshal(out.Bytes(), &r); err != nil || r.Decision != "PREPARED" || r.Capacity.InputJSONBytes != 1048576 || r.Capacity.NativeStdoutBytes != 1048576 || r.ProgramExecutions != 0 || r.ModelPredictions != 0 {
					t.Fatal(r, err)
				}
				return
			}
			raw := []byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"example/wire:First":0}]}`)
			size := 1048576
			if mode == "over" {
				size++
			}
			reader["inputs.json"] = append(raw, bytes.Repeat([]byte(" "), size-len(raw))...)
			suite, digest, inputOnly, err := readPackageReplayInputs(reader, map[string]string{"--inputs": "inputs.json"})
			if mode == "over" {
				if err == nil || !strings.Contains(err.Error(), "1048576 bytes") {
					t.Fatal(err)
				}
				return
			}
			if err != nil || !inputOnly || len(suite.Cases) != 1 || digest != workspaceDigest(reader["inputs.json"]) {
				t.Fatal(err, digest)
			}
		})
	}
}
