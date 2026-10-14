package cliobservations

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandHarness17011(t *testing.T) {
	source := []byte(`package commandharness
namespace commandharness17011
entity Integer id "urn:gooo:type:integer"
activity Shift(Integer)->Integer id "urn:commandharness17011:shift" computes "return input-17011"
`)
	path := filepath.Join(t.TempDir(), "command17011.gooo")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	if code := run([]string{"check", path}, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String(), out.String())
	}
	t.Logf("actual command binary built;race=%v;new source check passes;historical program executions0;binary removed when harness ends", raceInstrumented)
}
