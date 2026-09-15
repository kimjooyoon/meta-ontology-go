package query

func workspaceDatalogRequest() DatalogQuery {
	return DatalogQuery{
		Patterns: []Atom{Triple("dependsOn", Variable("entity"), Variable("source"))},
		Rules: []Rule{
			{
				ID:   "02/transitive-depends-on/v1",
				Head: Triple("dependsOn", Variable("entity"), Variable("source")),
				Body: []Atom{
					Triple("wasDerivedFrom", Variable("entity"), Variable("middle")),
					Triple("dependsOn", Variable("middle"), Variable("source")),
				},
			},
			{
				ID:   "01/direct-depends-on/v1",
				Head: Triple("dependsOn", Variable("entity"), Variable("source")),
				Body: []Atom{Triple("wasDerivedFrom", Variable("entity"), Variable("source"))},
			},
			{
				ID:   "00/inverse-used-by/v1",
				Head: Triple("usedBy", Variable("entity"), Variable("activity")),
				Body: []Atom{Triple("used", Variable("activity"), Variable("entity"))},
			},
		},
		IncludeDerived: true,
		Limit:          10,
	}
}
