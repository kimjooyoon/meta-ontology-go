package languageutility

import (
	"fmt"
	"strings"
	"unicode"
)

func GenerateProgram(contract Contract) (string, error) {
	if err := ValidateContract(contract); err != nil {
		return "", err
	}
	var source strings.Builder
	source.WriteString("entity LanguageUtilityCell id \"gooo://meta/language-utility/entity/cell\"\n")
	source.WriteString("entity LanguageUtilityEvidence id \"gooo://meta/language-utility/entity/evidence\"\n")
	source.WriteString("entity LanguageUtilityReport id \"gooo://meta/language-utility/entity/report\"\n\n")
	for _, useCase := range contract.UseCases {
		for _, stage := range contract.Stages {
			activity := "Observe" + symbol(useCase.ID) + symbol(stage.ID)
			computation := fmt.Sprintf("gooo.utility.%s.%s.%s:v1", useCase.ID,
				strings.ToLower(strings.ReplaceAll(stage.ID, "_", "-")), stage.ProofChoice)
			fmt.Fprintf(&source, "activity %s(LanguageUtilityCell) -> LanguageUtilityEvidence computes %q\n",
				activity, computation)
		}
	}
	source.WriteString("activity MeasureLanguageUtility(LanguageUtilityEvidence) -> LanguageUtilityReport computes \"gooo.metric.language-utility:v1\"\n")
	source.WriteString("activity ReplayLanguageUtility(LanguageUtilityReport) -> LanguageUtilityReport computes \"gooo.replay.language-utility:v1\"\n")
	return source.String(), nil
}

func symbol(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	var result strings.Builder
	for _, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}
	return result.String()
}
