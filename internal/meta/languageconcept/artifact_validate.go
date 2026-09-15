package languageconcept

import (
	"errors"
	"io/fs"
	"reflect"
)

func ValidateArtifact(repository fs.FS, artifact Artifact) error {
	if artifact.Schema != ArtifactSchema {
		return errors.New("language concept artifact schema mismatch")
	}
	if artifact.ArtifactDigest == "" || artifact.ArtifactDigest != artifactDigest(artifact) {
		return errors.New("language concept artifact digest mismatch")
	}
	expected := BuildArtifact(repository)
	if !reflect.DeepEqual(expected, artifact) {
		return errors.New("language concept artifact replay mismatch")
	}
	return nil
}
