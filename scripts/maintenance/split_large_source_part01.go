package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type options struct {
	root       string
	maxLines   int
	maxEntries int
	write      bool
	targets    []string
}
type declChunk struct {
	decls      []ast.Decl
	fileBodies []string
	declLines  int
	imports    map[string]struct{}
}
type importSpec struct {
	name string
	path string
}

func main() {
	root := flag.String("root", ".", "repository root for target files")
	maxLines := flag.Int("max-lines", sourcepolicy.DefaultMaxFileLines, "maximum lines per generated file")
	maxEntries := flag.Int("max-directory-entries", sourcepolicy.DefaultMaxDirectoryInputs, "maximum direct directory entries")
	write := flag.Bool("write", false, "write refactor files and remove originals")
	flag.Parse()
	if *maxLines <= 3 {
		fmt.Fprintln(os.Stderr, "max-lines must be greater than 3")
		os.Exit(1)
	}
	if *maxEntries <= 1 {
		fmt.Fprintln(os.Stderr, "max-directory-entries must be greater than 1")
		os.Exit(1)
	}
	opts := options{
		root:       *root,
		maxLines:   *maxLines,
		maxEntries: *maxEntries,
		write:      *write,
		targets:    flag.Args(),
	}
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
