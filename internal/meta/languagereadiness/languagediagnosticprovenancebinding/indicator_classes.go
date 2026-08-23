package languagediagnosticprovenancebinding

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"

func indicatorClasses(report languagediagnosticprovenance.Report) [3]int {
	result := [3]int{}
	for _, indicator := range report.Indicators {
		switch indicator.Class {
		case "OUTCOME":
			result[0]++
		case "DRIVER":
			result[1]++
		case "GUARDRAIL":
			result[2]++
		}
	}
	return result
}
