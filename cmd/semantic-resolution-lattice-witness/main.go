package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func main() {
	source := flag.String("source", "examples/semantic-resolution-lattice/main.gooo", "Gooo source")
	output := flag.String("output", "", "receipt output path; stdout when empty")
	flag.Parse()
	data, err := os.ReadFile(*source)
	if err != nil {
		fatal(err)
	}
	digest := sha256.Sum256(data)
	receipt := semanticresolution.BuildLatticeReceipt(*source, "sha256:"+hex.EncodeToString(digest[:]))
	if err := semanticresolution.ValidateLatticeReceipt(receipt); err != nil {
		fatal(err)
	}
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, err = os.Stdout.Write(encoded)
	} else {
		err = os.WriteFile(*output, encoded, 0o644)
	}
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
