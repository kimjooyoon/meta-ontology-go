package originpolicy

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo origin.generated.go
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo role.generated.go CandidateArtifactRole
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo subject.generated.go BindCandidateSubject
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo eligibility.generated.go ContinuationEligible
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo continuation_subject.generated.go ContinuationSubject
//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo continuation_name.generated.go ContinuationArtifactName

//go:embed source.gooo
var source []byte

func SourceDigest() string { return fmt.Sprintf("sha256:%x", sha256.Sum256(source)) }
