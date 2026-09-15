package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositorytopology"
)

func main() {
	root := flag.String("root", ".", "repository root")
	source := flag.String("source-metrics", "", "bound source metrics JSON")
	rootOntology := flag.String("root-ontology", "", "project-root ontology")
	bindingOntology := flag.String("binding-ontology", "", "meta-binding ontology")
	repository := flag.String("repository", "", "expected owner/repository")
	head := flag.String("expected-head", "", "expected exact head SHA")
	output := flag.String("output", "", "receipt path outside the repository")
	check := flag.Bool("check", false, "compare output with a fresh replay")
	flag.Parse()
	if *source == "" || *rootOntology == "" || *bindingOntology == "" || *output == "" {
		fatalf("source metrics, both ontologies, and output are required")
	}
	if !*check && inside(*root, *output) {
		fatalf("output must be outside the repository")
	}
	sourceJSON := read(*source)
	report, err := repositorytopology.Evaluate(sourceJSON, read(*rootOntology), read(*bindingOntology), *root, *repository, *head)
	if err != nil {
		fatalf("evaluate: %v", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatalf("encode: %v", err)
	}
	encoded = append(encoded, '\n')
	if *check {
		if !bytes.Equal(read(*output), encoded) {
			fatalf("receipt does not match replay")
		}
		return
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatalf("write receipt: %v", err)
	}
}

func read(name string) []byte {
	value, err := os.ReadFile(name)
	if err != nil {
		fatalf("read %s: %v", name, err)
	}
	return value
}

func inside(root, name string) bool {
	rootAbs, _ := filepath.Abs(root)
	nameAbs, _ := filepath.Abs(name)
	relative, err := filepath.Rel(rootAbs, nameAbs)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func fatalf(format string, values ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}
