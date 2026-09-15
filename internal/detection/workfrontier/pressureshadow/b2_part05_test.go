package pressureshadow

func positiveB2Result() B2Result {
	return B2Result{
		Schema:                           SchemaVersion,
		InputDigest:                      "sha256:a2517624fa346a2dad078c9b13d5f66ac6ca78b6ff75260003d82d0128cffc92",
		UpstreamResultDigest:             "sha256:cff0ff1a908890c9ab568d19a8d4a3c2a401781b1ae5aaf4e5effaf079a443f3",
		Decision:                         DecisionPass,
		Reason:                           ReasonNone,
		MissingRequiredPressureRecordIDs: []RequiredPressureSetIssue{},
		MissingSelectorPressureIDs:       []RequiredPressureSetIssue{},
		UnregisteredPressureRecordIDs:    []RequiredPressureSetIssue{},
		EnforcementEffect:                EnforcementNoEffect,
		ResultDigest:                     "sha256:16431ed57dd4f71911a15aa7f34ce13e326848a89d0dc28c0a077f62e5171c7d",
		ReplayDigest:                     "sha256:ade5d3c5ba3ad18be9aa95d207f5df525e31b446ec61f09f1fb604d19675380e",
	}
}
