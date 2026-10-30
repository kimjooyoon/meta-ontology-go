package generation

import (
	"crypto/sha256"
	_ "embed"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

//go:generate go run ../../../scripts/library-projection -root ../../.. -mode write -source internal/meta/generation/go-split-selection.gooo -out internal/meta/generation/go_split_selection.gooo.go -entry gooo://split-selection/select-next -go-package generation

//go:embed go-split-selection.gooo
var goSplitSelectionSource []byte

func goSplitSelectionSourceOrigin() sourcebundle.Member {
	return sourcebundle.Member{Filename: "internal/meta/generation/go-split-selection.gooo",
		SHA256: fmt.Sprintf("%x", sha256.Sum256(goSplitSelectionSource)), ByteLength: len(goSplitSelectionSource)}
}
