package policy

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo subjects.generated.go BindSubjects
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo eligibility.generated.go PermitContinuation
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo wait.generated.go PermitWait
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo target.generated.go BindTargetRun
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo stage.generated.go ObserveStage
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo artifact.generated.go ArtifactLabel
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo workflow.generated.go WorkflowPath
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo dependency.generated.go MissingDependency
//go:generate go run ../../../../scripts/transport-origin-codegen -sources source*.gooo next.generated.go NextObservation

//go:embed source*.gooo
var sources embed.FS

func policySources() ([]sourcebundle.Source, error) {
	names, err := fs.Glob(sources, "source*.gooo")
	if err != nil {
		return nil, err
	}
	return sourcebundle.Read(sources, names)
}

func SourceDigest() string {
	members, err := policySources()
	if err != nil {
		return ""
	}
	sum, err := sourcebundle.Digest(members)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("sha256:%x", sum)
}
