package rollbackintegrityeligibility

func denominatorContract() Suite {
	suite := Suite{Schema: SuiteSchema, DenominatorID: DenominatorID, Cases: make([]CaseResult, len(definitions))}
	for index, definition := range definitions {
		suite.Cases[index].Definition = definition
	}
	suite.DenominatorDigest = digestJSON(suite.Cases)
	return suite
}
