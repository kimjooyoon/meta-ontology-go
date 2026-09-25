package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"

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

func validateReady(repository fs.FS, artifact languageconcept.Artifact) error {
	if err := languageconcept.ValidateArtifact(repository, artifact); err != nil {
		return err
	}
	if artifact.Decision != "PASS" {
		return fmt.Errorf("FAIL_CLOSED: language concept artifact decision %q is not PASS", artifact.Decision)
	}
	if artifact.RepositoryWrites != 0 {
		return fmt.Errorf("FAIL_CLOSED: repository writes=%d", artifact.RepositoryWrites)
	}
	return nil
}

func printSummary(stdout io.Writer, artifact languageconcept.Artifact) {
	fmt.Fprintf(stdout, "language-concept-artifact: decision=%s paths=%d files=%d writes=%d digest=%s\n",
		artifact.Decision, artifact.Bindings.Paths, artifact.Bindings.Files,
		artifact.RepositoryWrites, artifact.ArtifactDigest)
}
