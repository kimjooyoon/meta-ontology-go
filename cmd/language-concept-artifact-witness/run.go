package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" {
		return fmt.Errorf("root is required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	artifactPath := cfg.output
	if cfg.check != "" {
		artifactPath = cfg.check
	}
	outside, err := outsideRoot(cfg.root, artifactPath)
	if err != nil {
		return err
	}
	if !outside {
		return fmt.Errorf("language concept artifact must be outside the repository root")
	}
	repository := os.DirFS(cfg.root)
	if cfg.check != "" {
		return consume(repository, cfg.check, stdout)
	}
	return produce(repository, cfg.output, stdout)
}

func produce(repository fs.FS, output string, stdout io.Writer) error {
	first := languageconcept.BuildArtifact(repository)
	if err := validateReady(repository, first); err != nil {
		return err
	}
	replay := languageconcept.BuildArtifact(repository)
	if err := validateReady(repository, replay); err != nil {
		return err
	}
	firstData, err := marshalArtifact(first)
	if err != nil {
		return err
	}
	replayData, err := marshalArtifact(replay)
	if err != nil {
		return err
	}
	if !bytes.Equal(firstData, replayData) {
		return fmt.Errorf("FAIL_CLOSED: language concept artifact replay diverged")
	}
	if err := writeExclusive(output, firstData); err != nil {
		return err
	}
	printSummary(stdout, first)
	return nil
}

func consume(repository fs.FS, path string, stdout io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	artifact := languageconcept.Artifact{}
	if err := json.Unmarshal(data, &artifact); err != nil {
		return err
	}
	if err := validateReady(repository, artifact); err != nil {
		return err
	}
	printSummary(stdout, artifact)
	return nil
}

func validateReady(repository fs.FS, artifact languageconcept.Artifact) error {
	if err := languageconcept.ValidateArtifact(repository, artifact); err != nil {
		return err
	}
	if artifact.Decision != "READY" {
		return fmt.Errorf("FAIL_CLOSED: language concept artifact decision %q is not READY", artifact.Decision)
	}
	if artifact.RepositoryWrites != 0 {
		return fmt.Errorf("FAIL_CLOSED: repository writes=%d", artifact.RepositoryWrites)
	}
	return nil
}

func marshalArtifact(artifact languageconcept.Artifact) ([]byte, error) {
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func outsideRoot(root, output string) (bool, error) {
	rootPath, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false, err
	}
	outputPath, err := filepath.Abs(filepath.Clean(output))
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return false, err
	}
	return relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)), nil
}

func printSummary(stdout io.Writer, artifact languageconcept.Artifact) {
	fmt.Fprintf(stdout, "language-concept-artifact: decision=%s paths=%d files=%d writes=%d digest=%s\n",
		artifact.Decision, artifact.Bindings.Paths, artifact.Bindings.Files,
		artifact.RepositoryWrites, artifact.ArtifactDigest)
}
