package extractor

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMapLiteralRuntimePreservesEffectsIdentityAndPanic(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("runtime witness requires CI=true")
	}
	original, generated := t.TempDir(), t.TempDir()
	support := map[string]string{"support.go": mapLiteralWitnessSupport()}
	if err := runtimeWitnessWriteModule(original, mapLiteralWitnessSource(), support); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(original, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWriteGoMod(generated); err != nil {
		t.Fatal(err)
	}
	if err := runtimeWitnessWriteSupport(generated, support); err != nil {
		t.Fatal(err)
	}
	for path, source := range result.Generated {
		if err := runtimeWitnessWriteRelative(generated, path, source); err != nil {
			t.Fatal(err)
		}
	}
	before, stderr, err := runtimeWitnessRunGo(original)
	if err != nil {
		t.Fatalf("original: %v\n%s", err, stderr)
	}
	after, stderr, err := runtimeWitnessRunGo(generated)
	if err != nil {
		t.Fatalf("generated: %v\n%s", err, stderr)
	}
	if string(before) != mapLiteralExpectedOutput() || string(after) != string(before) {
		t.Fatalf("before=%q after=%q expected=%q", before, after, mapLiteralExpectedOutput())
	}
}

func mapLiteralWitnessSource() string {
	var source strings.Builder
	source.WriteString("package main\n\nfunc Run(next func(string) any) map[string]any {\n")
	source.WriteString(strings.Repeat("\t_ = 1\n", 60))
	source.WriteString("\twitness := map[string]any{\n")
	for index := 1; index <= 16; index++ {
		fmt.Fprintf(&source, "\t\t%q: next(%q),\n", fmt.Sprintf("%02d", index), fmt.Sprintf("%02d", index))
	}
	source.WriteString("\t}\n\tpublish(witness)\n\treturn witness\n}\n")
	return source.String()
}

func mapLiteralExpectedOutput() string {
	values := make([]string, 0, 17)
	for index := 1; index <= 16; index++ {
		values = append(values, fmt.Sprintf("%02d@main.Run", index))
	}
	values = append(values, "publish")
	return strings.Join(values, ",") + "\ncount=16 nil=true typed_nil=false integer=int\npanic=stop events=01@main.Run,02@main.Run\n"
}

func mapLiteralWitnessSupport() string {
	return `package main

import (
	"fmt"
	"runtime"
	"strings"
)

var events []string
var failAt string

func next(key string) any {
	pc, _, _, _ := runtime.Caller(1)
	events = append(events, key+"@"+runtime.FuncForPC(pc).Name())
	if key == failAt {
		panic("stop")
	}
	switch key {
	case "01":
		return 1
	case "02":
		return (*int)(nil)
	case "03":
		return nil
	default:
		return key
	}
}

func publish(value map[string]any) { events = append(events, "publish") }

func runPanic() {
	defer func() {
		fmt.Printf("panic=%v events=%s\n", recover(), strings.Join(events, ","))
	}()
	Run(next)
}

func main() {
	value := Run(next)
	fmt.Println(strings.Join(events, ","))
	fmt.Printf("count=%d nil=%t typed_nil=%t integer=%T\n", len(value), value["03"] == nil, value["02"] == nil, value["01"])
	events = nil
	failAt = "02"
	runPanic()
}
`
}
