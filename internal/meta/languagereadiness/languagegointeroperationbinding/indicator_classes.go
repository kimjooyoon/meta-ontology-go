package languagegointeroperationbinding

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagegointeroperation"

func indicatorClasses(report languagegointeroperation.Report) [3]int {
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
