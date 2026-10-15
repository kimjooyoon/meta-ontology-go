package languagesyntax

func expectedSyntaxCases() []CaseDefinition {
	cases := expectedSyntaxCaseGroup0()
	cases = append(cases, expectedSyntaxCaseGroup1()...)
	cases = append(cases, expectedSyntaxCaseGroup2()...)
	cases = append(cases, expectedSyntaxCaseGroup3()...)
	cases = append(cases, expectedSyntaxCaseGroup4()...)
	cases = append(cases, expectedSyntaxCaseGroup5()...)
	return cases
}
