package policy

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo subjects.generated.go BindSubjects
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo eligibility.generated.go PermitContinuation
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo wait.generated.go PermitWait
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo target.generated.go BindTargetRun
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo stage.generated.go ObserveStage
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo artifact.generated.go ArtifactLabel
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo workflow.generated.go WorkflowPath
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo dependency.generated.go MissingDependency
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo next.generated.go NextObservation

//go:embed source.gooo
var source []byte

func SourceDigest() string { return fmt.Sprintf("sha256:%x", sha256.Sum256(source)) }
