package bodylibrary

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestZeroInputStageExportsAsATypedNativeConstant(t *testing.T) {
	source := []byte(`package constants
namespace constants
entity Integer id "urn:gooo:type:integer"
activity Seed() -> Integer id "urn:zero:seed" computes "return 7"
activity Main(Integer, Integer) -> Integer id "urn:zero:main" computes "return input0 + input1"
bind Seed.result -> Main.input0
`)
	generated, err := GenerateWithStageExports(context.Background(), "zero.gooo", source, "Main", "constant",
		[]StageExport{{Name: "Seed", ActivityID: "urn:zero:seed"}})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "constant"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"go.mod":                    "module example.test/constant\ngo 1.27.3\n",
		"constant/semantic.gooo.go": generated,
		"consumer_test.go": `package consumer
import("testing"; p "example.test/constant/constant")
func TestValues(t *testing.T){
 if p.Seed()!=7 {t.Fatal("source constant")}
 for _,row:=range []struct{in,want int64}{{-7,0},{0,7},{5,12}} {
  if got:=p.Evaluate(row.in);got!=row.want {t.Fatal(row,got)}
 }
}`,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./...")
	command.Dir, command.Env = root, append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
}
