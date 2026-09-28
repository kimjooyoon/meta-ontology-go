package artifactresolutionexperiment

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func goldenEqual(actual, expected artifactemit.Artifact) bool {
	actual.SubjectDigest = ""
	actual.Digest = ""
	expected.SubjectDigest = ""
	expected.Digest = ""
	return reflect.DeepEqual(actual, expected)
}
