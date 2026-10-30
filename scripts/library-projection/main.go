package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
)

const sourcePath = "internal/meta/releaseidentity/version.gooo"
const outputPath = "internal/meta/releaseidentity/semantic.gooo.go"
const activityID = "release://activity/current-version"

type projectionTarget struct {
	source, output, activity, goPackage string
}

func main() {
	root := flag.String("root", ".", "repository root")
	mode := flag.String("mode", "check", "check or write a source-owned Go library projection")
	source := flag.String("source", "", "Gooo source path; supply all four target flags together")
	output := flag.String("out", "", "generated Go output path")
	entry := flag.String("entry", "", "stable activity ID")
	goPackage := flag.String("go-package", "", "generated Go package")
	flag.Parse()
	if flag.NArg() != 0 || (*mode != "check" && *mode != "write") {
		fail(fmt.Errorf("expected -mode check|write and no positional arguments"))
	}
	target, err := selectedTarget(*source, *output, *entry, *goPackage)
	if err != nil {
		fail(err)
	}
	if err := projectLibrary(*root, target, *mode == "write"); err != nil {
		fail(err)
	}
}

func selectedTarget(source, output, entry, goPackage string) (projectionTarget, error) {
	if source == "" && output == "" && entry == "" && goPackage == "" {
		return projectionTarget{sourcePath, outputPath, activityID, "releaseidentity"}, nil
	}
	if source == "" || output == "" || entry == "" || goPackage == "" {
		return projectionTarget{}, fmt.Errorf("source, out, entry and go-package must be supplied together")
	}
	return projectionTarget{source, output, entry, goPackage}, nil
}

func project(root string, write bool) error {
	return projectLibrary(root, projectionTarget{sourcePath, outputPath, activityID, "releaseidentity"}, write)
}

func projectLibrary(root string, target projectionTarget, write bool) error {
	sourceFile, err := filepath.Abs(filepath.Join(root, target.source))
	if err != nil {
		return err
	}
	path, err := filepath.Abs(filepath.Join(root, target.output))
	if err != nil {
		return err
	}
	if sourceFile == path {
		return fmt.Errorf("library projection output must differ from its Gooo source")
	}
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	generated, err := bodylibrary.Generate(context.Background(), target.source, source, target.activity, target.goPackage)
	if err != nil {
		return err
	}
	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, []byte(generated)) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !write {
		return fmt.Errorf("library projection %s differs from Gooo source: run go generate for its owning package", target.output)
	}
	return publish(path, []byte(generated))
}

func publish(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".library-projection-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Chmod(0644); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
