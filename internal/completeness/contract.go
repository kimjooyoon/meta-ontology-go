// Package completeness contains the source-owned receipt structure used by
// compiler producers and independent runtime/reverse observation consumers.
package completeness

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
)

//go:generate go run ../../scripts/receipt-schema --source receipt.gooo --go receipt.generated.go --json receipt.schema.json
//go:embed receipt.gooo
var declaration []byte

var declarationSHA256 = fmt.Sprintf("%x", sha256.Sum256(declaration))

func DeclarationSource() []byte { return bytes.Clone(declaration) }

func DeclarationMatchesProjection() bool { return declarationSHA256 == DeclarationSHA256 }

func ContractBinding() map[string]string {
	return map[string]string{"schema": CompletenessReceiptSchema, "declaration_sha256": declarationSHA256, "generated_declaration_sha256": DeclarationSHA256, "projection_profile": ProjectionProfile}
}
