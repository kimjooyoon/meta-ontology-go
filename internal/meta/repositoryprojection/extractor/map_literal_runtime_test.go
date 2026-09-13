package extractor

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMapLiteralRuntimePreservesEffectsIdentityAndPanic(t *testing.T) {
	assertMapLiteralRuntimeWitness(t, mapLiteralWitnessSource(), mapLiteralWitnessSupport(), mapLiteralExpectedOutput())
}

func TestMapReceiverFusionRuntimePreservesCallerAndOrder(t *testing.T) {
	assertMapLiteralRuntimeWitness(t, mapReceiverWitnessSource(), mapReceiverWitnessSupport(), mapReceiverExpectedOutput())
}

func assertMapLiteralRuntimeWitness(t *testing.T, source, supportSource, expected string) {
	t.Helper()
	if os.Getenv("CI") != "true" {
		t.Skip("runtime witness requires CI=true")
	}
	original, generated := t.TempDir(), t.TempDir()
	support := map[string]string{"support.go": supportSource}
	if err := runtimeWitnessWriteModule(original, source, support); err != nil {
		t.Fatal(err)
	}
	result, err := ExtractWithResult(original, "x.go")
	if err != nil {
		t.Fatalf("extract runtime witness: %#v", err)
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
	if string(before) != expected || string(after) != string(before) {
		t.Fatalf("before=%q after=%q expected=%q", before, after, expected)
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

func mapReceiverWitnessSource() string {
	source := strings.Replace(mapLiteralWitnessSource(), strings.Repeat("\t_ = 1\n", 60), strings.Repeat("\t_ = 1\n", 64), 1)
	preparation := "\tcommand := receiverFactory()\n" +
		"\tvalue, failure := command.consume(receiverArgument())\n" +
		"\tif value != \"ok\" || failure != nil {\n\t\tpanic(\"bad receiver result\")\n\t}\n"
	return strings.Replace(source, "\twitness := map[string]any{", preparation+"\twitness := map[string]any{", 1)
}

func mapReceiverWitnessSupport() string {
	base := mapLiteralWitnessSupport()
	return base[:strings.Index(base, "func main()")] + mapReceiverRuntimeTypes() + mapReceiverRuntimeMain()
}

func mapReceiverRuntimeTypes() string {
	return `
var receiverPhase string
type receiver struct{}

func receiverEvent(label string) {
	pc, _, _, _ := runtime.Caller(2)
	events = append(events, label+"@"+runtime.FuncForPC(pc).Name())
	if receiverPhase == label { panic("stop") }
}
func receiverFactory() *receiver {
	receiverEvent("init")
	if receiverPhase == "nil" { return nil }
	return &receiver{}
}
func receiverArgument() int {
	receiverEvent("argument")
	return 7
}
func (r *receiver) consume(value int) (string, error) {
	receiverEvent("method")
	if value != 7 || (r == nil) != (receiverPhase == "nil") { panic("bad receiver identity or argument") }
	return "ok", nil
}
`
}

func mapReceiverRuntimeMain() string {
	return `
func receiverScenario(mode string) {
	receiverPhase, events, failAt = mode, nil, ""
	if mode == "map" { failAt = "02" }
	defer func() {
		fmt.Printf("%s panic=%v events=%s\n", mode, recover(), strings.Join(events, ","))
	}()
	value := Run(next)
	fmt.Printf("%s count=%d nil=%t typed_nil=%t integer=%T\n", mode, len(value), value["03"] == nil, value["02"] == nil, value["01"])
}
func main() {
	for _, mode := range []string{"normal", "nil", "init", "argument", "method", "map"} {
		receiverScenario(mode)
	}
}
`
}

func mapReceiverExpectedOutput() string {
	prefix := []string{"init@main.Run", "argument@main.Run", "method@main.Run"}
	normal := append([]string{}, prefix...)
	for index := 1; index <= 16; index++ {
		normal = append(normal, fmt.Sprintf("%02d@main.Run", index))
	}
	normal = append(normal, "publish")
	var expected strings.Builder
	for _, mode := range []string{"normal", "nil"} {
		fmt.Fprintf(&expected, "%s count=16 nil=true typed_nil=false integer=int\n", mode)
		fmt.Fprintf(&expected, "%s panic=<nil> events=%s\n", mode, strings.Join(normal, ","))
	}
	for index, mode := range []string{"init", "argument", "method"} {
		fmt.Fprintf(&expected, "%s panic=stop events=%s\n", mode, strings.Join(prefix[:index+1], ","))
	}
	fmt.Fprintf(&expected, "map panic=stop events=%s,01@main.Run,02@main.Run\n", strings.Join(prefix, ","))
	return expected.String()
}
