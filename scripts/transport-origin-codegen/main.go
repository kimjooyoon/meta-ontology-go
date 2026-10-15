// Command transport-origin-codegen projects the source-owned origin relation.
package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		return fmt.Errorf("usage: transport-origin-codegen source.gooo output.go [activity]")
	}
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	activity := "BindProducerRun"
	if len(os.Args) == 4 {
		activity = os.Args[3]
	}
	result, err := bodycodegen.Generate(os.Args[1], source, activity)
	if err != nil {
		return err
	}
	return os.WriteFile(os.Args[2], []byte(result.Source), 0644)
}
