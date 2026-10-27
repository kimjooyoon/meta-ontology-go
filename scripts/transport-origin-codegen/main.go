// Command transport-origin-codegen projects the source-owned origin relation.
package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	return generateArgs(os.Args[1:])
}

func generateArgs(args []string) error {
	bundle := len(args) > 0 && args[0] == "-sources"
	if bundle {
		args = args[1:]
	}
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("usage: transport-origin-codegen [-sources] source.gooo-or-pattern output.go [activity]")
	}
	activity := "BindProducerRun"
	if len(args) == 3 {
		activity = args[2]
	}
	result, err := projection(args[0], activity, bundle)
	if err != nil {
		return err
	}
	return os.WriteFile(args[1], []byte(result), 0644)
}

func projection(name, activity string, bundle bool) (string, error) {
	if bundle {
		return bundleProjection(os.DirFS("."), name, activity)
	}
	source, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	result, err := bodycodegen.Generate(name, source, activity)
	return result.Source, err
}

func bundleProjection(fsys fs.FS, pattern, activity string) (string, error) {
	names, err := fs.Glob(fsys, pattern)
	if err != nil {
		return "", err
	}
	sources, err := sourcebundle.Read(fsys, names)
	if err != nil {
		return "", err
	}
	result, err := bodycodegen.GenerateSourceBundle(sources, activity)
	return result.Source, err
}
