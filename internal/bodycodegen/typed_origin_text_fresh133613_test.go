package bodycodegen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const readableSource133613 = "package readable\nnamespace readable\nentity Integer id \"urn:gooo:type:integer\"\nentity Boolean id \"urn:gooo:type:boolean\"\n" +
	"# 원래 소스의 위치를 유지한다\nactivity Revise(Integer,Integer,Boolean) -> Integer id \"urn:readable133613:revise\" computes `\nlet subtotal=input0\nsubtotal+=input1\nif input2 {subtotal*=3}else{subtotal-=17}\nreturn subtotal\n`\n"

func TestReadableSourceOrigins133613(t *testing.T) {
	source := []byte(readableSource133613)
	report, err := InspectTypedValueOrigins("readable133613.gooo", source, "Revise")
	if err != nil || report.Status != "RESOLVED" {
		t.Fatal(err, report.Reason)
	}
	before, _ := json.Marshal(report)
	var text bytes.Buffer
	if err = WriteTypedOriginText(&text, report, source); err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{"urn:readable133613:revise", "Scope: possible source dependencies", "input0", "input1", "input2", "subtotal", "data <- #", "control <- #", "possible value merge", "at 8:"} {
		if !strings.Contains(text.String(), wanted) {
			t.Fatalf("missing %q in %s", wanted, text.String())
		}
	}
	after, _ := json.Marshal(report)
	if !bytes.Equal(before, after) {
		t.Fatal("display changed source graph")
	}
	if err = WriteTypedOriginText(&bytes.Buffer{}, report, append(source, '\n')); err == nil {
		t.Fatal("display accepted different source bytes")
	}
	broken := report
	broken.Sources = nil
	if err = WriteTypedOriginText(&bytes.Buffer{}, broken, source); err == nil {
		t.Fatal("missing source extent accepted")
	}
}

func TestGoooRoleProjection133613(t *testing.T) {
	var origin struct{ Source, Generated string }
	raw, err := os.ReadFile("readable_origins/generation-origin.json")
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]string
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	origin.Source, origin.Generated = fields["source_sha256"], fields["generated_sha256"]
	for name, wanted := range map[string]string{"readable_origins/roles.gooo": origin.Source, "origin_role_generated.go": origin.Generated} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != wanted {
			t.Fatal("generation origin differs", name)
		}
	}
	if OriginRole("copy") != "local definition" || OriginRole("join") != "possible value merge" || OriginRole("future133613") != "unclassified source role" {
		t.Fatal("source display policy differs")
	}
}
