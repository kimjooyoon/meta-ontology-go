package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func main() {
	root := flag.String("root", "", "read-only repository root")
	logical := flag.String("logical", "cmd/language-readiness-witness/predecessor-selection/pagination_test.go", "logical pagination fixture path")
	output := flag.String("output", "", "caller-owned preview JSON output")
	flag.Parse()
	if *root == "" || *output == "" {
		fatal("root and output are required")
	}
	preview, err := extractor.PreviewBoundedPaginationCallback(*root, filepath.ToSlash(*logical))
	if err != nil {
		fatal(err.Error())
	}
	payload, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		fatal(err.Error())
	}
	if err := os.WriteFile(filepath.Clean(*output), append(payload, '\n'), 0o644); err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
