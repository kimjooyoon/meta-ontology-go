package sourceauthorityshadow

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityeval"
)

func Observe(raw []byte, expectedSHA string) Receipt {
	evaluation := sourceauthorityeval.Observe(raw)
	receipt := Receipt{
		Schema: ReceiptSchema, Mode: Mode, ExpectedSHA: expectedSHA,
		SubjectSHA: evaluation.SubjectSHA, InputDigest: sourceauthorityeval.DigestBytes(raw),
		Observation: evaluation.Observation, Resolution: evaluation.Resolution,
		Enforcement: evaluation.Enforcement, Reason: evaluation.Reason,
		GateEffect: "NO_EFFECT", Evaluation: evaluation, Indicators: []Indicator{},
	}
	headBound := exactSHA(expectedSHA) && evaluation.SubjectSHA == expectedSHA
	if !headBound {
		receipt.Observation = "UNKNOWN"
		receipt.Resolution = "INVARIANT_ONLY"
		receipt.Enforcement = "BLOCK"
		receipt.Reason = "SOURCE_AUTHORITY_SHADOW_HEAD_UNKNOWN"
	}
	receipt.Indicators = buildIndicators(receipt, headBound)
	return seal(receipt)
}

func exactSHA(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
