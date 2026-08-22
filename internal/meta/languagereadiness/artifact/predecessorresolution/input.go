package predecessorresolution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

type Attempt struct {
	Depth       int
	AncestorSHA string
	ParentSHA   string
	Selection   predecessorselection.Result
}

type Input struct {
	Repository              string
	CurrentHeadSHA           string
	ImmediatePredecessorSHA  string
	SearchLimit              int
	Attempts                 []Attempt
}
