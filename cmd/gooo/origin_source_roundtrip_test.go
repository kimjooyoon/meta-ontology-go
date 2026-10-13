package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalCandidateSourceRoundtrip5719(t *testing.T) {
	const source = `package roundorigin
namespace roundorigin5719
entity Integer id "urn:gooo:type:integer"
activity Scale(Integer) -> Integer id "urn:roundorigin5719:scale" computes "let shouldScale = input > __GOOO_BODY_HOLE_boundary__; let amount = input; if shouldScale { amount = input * 5 }; return amount" assembling {
 source_fill intent "선택한 경계를 변수에 담아 배율 적용 여부를 정한다" {
  hole "boundary"
  candidate "boundary5719" { fill "boundary" "79" }
  candidate "other5719" { fill "boundary" "179" }
 }
 case "83" -> "415"
}
`
	path := filepath.Join(t.TempDir(), "source.gooo")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"roundtrip", "--json", path}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte(`"get_put":true`)) || !bytes.Contains(output.Bytes(), []byte(`"put_get":true`)) {
		t.Fatal("roundtrip laws absent", output.String())
	}
	if target := os.Getenv("GOOO_ORIGIN_QUERY_OUTPUT"); target != "" {
		if err := os.WriteFile(filepath.Join(target, "roundtrip5719.json"), output.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("5719: fresh local-candidate source passes semantic roundtrip, Get-Put and Put-Get; no body or model execution")
}
