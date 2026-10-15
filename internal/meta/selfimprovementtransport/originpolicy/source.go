package originpolicy

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../../../scripts/transport-origin-codegen source.gooo origin.generated.go

//go:embed source.gooo
var source []byte

func SourceDigest() string { return fmt.Sprintf("sha256:%x", sha256.Sum256(source)) }
