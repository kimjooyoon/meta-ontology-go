// Package completenessdelta compares source-bound observations without executing
// code, loading models, changing intent, or turning unknown evidence into zero.
package completenessdelta

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../scripts/receipt-schema --source delta.gooo --root CompletenessDelta --go delta.generated.go --json delta.schema.json
//go:embed delta.gooo
var declaration []byte

func digest(data []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(data)) }

func DeclarationSource() []byte { return bytes.Clone(declaration) }

func declarationMatches() bool { return digest(declaration) == "sha256:"+DeclarationSHA256 }
