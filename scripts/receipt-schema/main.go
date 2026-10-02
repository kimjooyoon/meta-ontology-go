// Command receipt-schema regenerates the source-owned compiler receipt DTO and
// its JSON schema. It is a deterministic local build tool without model calls.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func main() {
	source := flag.String("source", "", "Gooo declaration")
	goOutput := flag.String("go", "", "generated Go output")
	jsonOutput := flag.String("json", "", "JSON schema output")
	flag.Parse()
	if *source == "" || *goOutput == "" || *jsonOutput == "" || flag.NArg() != 0 {
		fail(fmt.Errorf("source, go and json paths are required"))
	}
	in, err := os.Open(*source)
	if err != nil {
		fail(err)
	}
	defer in.Close()
	raw, err := io.ReadAll(io.LimitReader(in, 64<<10+1))
	if err != nil {
		fail(err)
	}
	p, err := receiptprojection.Compile(*source, raw, "CompletenessReceipt")
	if err != nil {
		fail(err)
	}
	goBytes, err := p.Go()
	if err != nil {
		fail(err)
	}
	jsonBytes, err := p.JSONSchema()
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(*goOutput, goBytes, 0644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*jsonOutput, jsonBytes, 0644); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
