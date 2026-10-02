package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func runReceiptSchema(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("receipt-schema", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "json", "go or json")
	root := flags.String("root", "CompletenessReceipt", "root receipt entity")
	if err := flags.Parse(args); err != nil || flags.NArg() != 1 || (*format != "go" && *format != "json") {
		fmt.Fprintln(stderr, "usage: gooo receipt-schema [--format go|json] [--root Entity] source.gooo")
		return exitUsage
	}
	f, e := os.Open(flags.Arg(0))
	if e != nil {
		fmt.Fprintln(stderr, e)
		return exitFailure
	}
	defer f.Close()
	source, e := io.ReadAll(io.LimitReader(f, 64<<10+1))
	if e != nil {
		fmt.Fprintln(stderr, e)
		return exitFailure
	}
	p, e := receiptprojection.Compile(flags.Arg(0), source, *root)
	if e != nil {
		fmt.Fprintln(stderr, e)
		return exitFailure
	}
	var out []byte
	if *format == "go" {
		out, e = p.Go()
	} else {
		out, e = p.JSONSchema()
	}
	if e != nil {
		fmt.Fprintln(stderr, e)
		return exitFailure
	}
	if _, e = stdout.Write(out); e != nil {
		return exitFailure
	}
	return exitOK
}
