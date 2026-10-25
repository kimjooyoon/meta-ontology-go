package completion

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../transport-origin-codegen source.gooo classify.generated.go Classify
//go:generate go run ../../transport-origin-codegen source.gooo decision.generated.go Decision
//go:generate go run ../../transport-origin-codegen source.gooo resolution.generated.go Resolution
//go:generate go run ../../transport-origin-codegen source.gooo next.generated.go NextOperation

//go:embed source.gooo
var source []byte

func SourceDigest() string { return fmt.Sprintf("sha256:%x", sha256.Sum256(source)) }
