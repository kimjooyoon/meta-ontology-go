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

func main() {
	root := flag.String("root", ".", "repository root")
	mode := flag.String("mode", "check", "check or write the source-owned release version projection")
	flag.Parse()
	if flag.NArg() != 0 || (*mode != "check" && *mode != "write") {
		fail(fmt.Errorf("expected -mode check|write and no positional arguments"))
	}
	if err := project(*root, *mode == "write"); err != nil {
		fail(err)
	}
}

func project(root string, write bool) error {
	source, err := os.ReadFile(filepath.Join(root, sourcePath))
	if err != nil {
		return err
	}
	generated, err := bodylibrary.Generate(context.Background(), sourcePath, source, activityID, "releaseidentity")
	if err != nil {
		return err
	}
	path := filepath.Join(root, outputPath)
	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, []byte(generated)) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !write {
		return fmt.Errorf("release version projection differs from Gooo source: run go generate ./internal/meta/releaseidentity")
	}
	return publish(path, []byte(generated))
}

func publish(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".release-version-*.tmp")
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
