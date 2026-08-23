package languagereadiness

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

type PromotionEvidence struct {
	Promotion      proposalpromotion.Receipt
	Capability     guardedcapability.Receipt
	UseCases       toolchainusecases.Report
	Syntax         languagesyntax.Report
	Diagnostic     languagediagnosticprovenance.Report
	PackageRuntime []languagepackageruntime.Report
}
