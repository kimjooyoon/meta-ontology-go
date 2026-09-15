package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

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
