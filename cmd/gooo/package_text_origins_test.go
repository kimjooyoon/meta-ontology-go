package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageTextOriginalBytesAndDecodedValue(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := "../../examples/text-input-origins"
	var actual []byte
	var previousDigest string
	for _, name := range []string{"raw-input.json", "escaped-input.json"} {
		input, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		var out, diag bytes.Buffer
		code := run([]string{"package", "execute", "--json", "--inputs", filepath.Join(root, name), filepath.Join(root, "gooo.workspace.json")}, &out, &diag)
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
			t.Fatal(name, code, err, receipt.Error, diag.String())
		}
		if receipt.InputsDigest != workspaceDigest(input) || receipt.InputsDigest == previousDigest || receipt.CasesDigest != "" || receipt.Decision != "OBSERVED" {
			t.Fatal("original JSON identity changed", name, receipt.InputsDigest)
		}
		runtime := receipt.Result.Runtime
		if runtime.ModelCalls != 0 || runtime.FiniteTotal != 0 || len(runtime.Traces) != 1 {
			t.Fatal("input-only source scope changed", name)
		}
		value := runtime.Traces[0].Deliveries[0].Actual
		var description struct {
			Text  string `json:"text"`
			Bytes int    `json:"utf8_bytes"`
		}
		if err := json.Unmarshal(value, &description); err != nil || description.Text != "기원:🧭�" || description.Bytes != 14 {
			t.Fatal("decoded Text differs", name, description, err)
		}
		if actual != nil && !bytes.Equal(actual, value) {
			t.Fatal("equivalent Unicode spellings changed the value")
		}
		actual, previousDigest = append([]byte(nil), value...), receipt.InputsDigest
	}
}

func TestPackageTextRejectsMalformedEscapesBeforeModel(t *testing.T) {
	manifest := "../../examples/text-input-origins/gooo.workspace.json"
	for _, c := range []struct{ value, reason string }{
		{`"prefix-\ud809"`, "UNPAIRED_HIGH"},
		{`"prefix-\udffe"`, "UNPAIRED_LOW"},
		{`"\ud80a\\udc02"`, "UNPAIRED_HIGH"},
		{string([]byte{'"', 'p', 0xff, 'q', '"'}), "UTF-8"},
	} {
		input := []byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"origin/text:Describe":` + c.value + `}]}`)
		file := filepath.Join(t.TempDir(), "original-input.json")
		if err := os.WriteFile(file, input, 0600); err != nil {
			t.Fatal(err)
		}
		var out, diag bytes.Buffer
		code := run([]string{"package", "execute", "--json", "--inputs", file, "--tiny-model", "unread-text-test-model.json", manifest}, &out, &diag)
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code == exitOK || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, c.reason) {
			t.Fatal("malformed Text reached later processing", c.reason, code, err, receipt.Error)
		}
		if receipt.InputsDigest != workspaceDigest(input) || receipt.Result != nil || strings.Contains(receipt.Error, "model could not be loaded") {
			t.Fatal("rejected original input lost identity or loaded the model", receipt.InputsDigest, receipt.Error)
		}
	}
}
