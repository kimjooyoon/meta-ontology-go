package bodyexecution

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestTextInputEscapePolicyBoundaries(t *testing.T) {
	for _, c := range []struct {
		unit, next int64
		want       string
	}{
		{-1, -1, "INVALID_UNIT"}, {65536, -1, "INVALID_UNIT"},
		{0, -1, "SCALAR_ESCAPE"}, {55295, -1, "SCALAR_ESCAPE"},
		{55296, 56320, "SURROGATE_PAIR"}, {56319, 57343, "SURROGATE_PAIR"},
		{55297, -1, "UNPAIRED_HIGH"}, {56318, 57344, "UNPAIRED_HIGH"},
		{56320, 55296, "UNPAIRED_LOW"}, {57343, -1, "UNPAIRED_LOW"},
		{57344, -1, "SCALAR_ESCAPE"}, {65535, -1, "SCALAR_ESCAPE"},
	} {
		if got := TextEscapePolicy(c.unit, c.next); got != c.want {
			t.Fatalf("unit=%d next=%d: %s, want %s", c.unit, c.next, got, c.want)
		}
	}
}

func TestTextInputJSONScalarPairs(t *testing.T) {
	for _, c := range []struct{ raw, want, reason string }{
		{`"기원:\uD83e\uDdEd\ufffd"`, "기원:🧭�", ""},
		{`"literal\\ud802"`, `literal\ud802`, ""},
		{`"quote:\";literal:\\;nul:\u0000"`, "quote:\";literal:\\;nul:\x00", ""},
		{`"\ud7ff\ue000"`, "\ud7ff\ue000", ""},
		{`"prefix-\ud801-suffix"`, "", "UNPAIRED_HIGH"},
		{`"prefix-\udfff-suffix"`, "", "UNPAIRED_LOW"},
		{`"\ud802\ud803"`, "", "UNPAIRED_HIGH"},
		{`"\ud804\\udc01"`, "", "UNPAIRED_HIGH"},
		{`"literal\\ud805;actual:\ud806"`, "", "UNPAIRED_HIGH"},
	} {
		got, err := decodeTextScalar([]byte(c.raw))
		if c.reason != "" {
			if err == nil || !strings.Contains(err.Error(), c.reason) {
				t.Fatal(c.raw, "expected", c.reason, "got", err)
			}
		} else if err != nil || got != c.want {
			t.Fatal(c.raw, got, err)
		}
	}
	for _, raw := range [][]byte{[]byte(`null`), []byte(`"\uZZZZ"`), {'"', 0xff, '"'}, []byte(`"` + strings.Repeat("🧭", 257) + `"`)} {
		if _, err := decodeTextScalar(raw); err == nil {
			t.Fatal("invalid or oversized Text accepted")
		}
	}
	if _, err := decodeTextScalar([]byte(`"` + strings.Repeat("🧭", 256) + `"`)); err != nil {
		t.Fatal("1024 decoded UTF-8 bytes refused", err)
	}
}

func TestTextInputJSONCompositionBoundaries(t *testing.T) {
	bad := `"suffix-\ud807"`
	inputs := []byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"Run":` + bad + `}]}`)
	cases := []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Run":"ok"},"expected":{"Run":` + bad + `}}]}`)
	series := []byte(`{"schema":"gooo/body-composition-case-series/v1","suites":[` + string(cases) + `]}`)
	for _, c := range []struct {
		raw []byte
		err error
	}{
		{inputs, textInputDecodeError(inputs, "inputs")},
		{cases, textInputDecodeError(cases, "cases")},
		{series, textInputDecodeError(series, "series")},
	} {
		if c.err == nil || !strings.Contains(c.err.Error(), "UNPAIRED_HIGH") {
			t.Fatal("escaped input normalized", c.err)
		}
		wantOffset := strings.Index(string(c.raw), `\ud807`)
		if !strings.Contains(c.err.Error(), "byte "+strconv.Itoa(wantOffset)+":") {
			t.Fatal("error lost original JSON byte offset", c.err, wantOffset)
		}
	}
	if _, err := canonicalRecordField([]byte(`["ok","\ud808"]`), "Text", "many"); err == nil {
		t.Fatal("programmatic repeated Text bypassed escape policy")
	}
	canonical, err := canonicalScalar([]byte(`"기원:\ud83e\udded\ufffd"`), "Text")
	if err != nil || !bytes.Equal(canonical, []byte(`"기원:🧭�"`)) {
		t.Fatal("valid decoded identity differs", string(canonical), err)
	}
}

func textInputDecodeError(raw []byte, kind string) error {
	switch kind {
	case "inputs":
		_, err := DecodeCompositionInputs(raw)
		return err
	case "cases":
		_, err := DecodeCompositionCases(raw)
		return err
	default:
		_, err := DecodeCompositionCaseSeries(raw)
		return err
	}
}

func TestTextInputPolicyGenerationOrigins(t *testing.T) {
	raw, err := os.ReadFile("../meta/textinput/generation-origin.json")
	if err != nil {
		t.Fatal(err)
	}
	var origin map[string]string
	if err := json.Unmarshal(raw, &origin); err != nil || origin["decision"] != "PASS" {
		t.Fatal("policy generation origin", err)
	}
	for file, key := range map[string]string{
		"../meta/textinput/escapes.gooo":          "source_sha256",
		"../valuejson/escape_policy_generated.go": "generated_sha256",
	} {
		data, err := os.ReadFile(file)
		if key == "generated_sha256" {
			// Only the handwritten host package declaration moved. Compare the
			// recovered original bytes against the unchanged historical receipt.
			if !bytes.HasPrefix(data, []byte("package valuejson\n")) {
				t.Fatal("unexpected relocated projection package", file)
			}
			data = bytes.Replace(data, []byte("package valuejson\n"), []byte("package bodyexecution\n"), 1)
		}
		digest := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(digest[:]) != origin[key] {
			t.Fatal("policy source/projection identity differs", file, err)
		}
	}
}
