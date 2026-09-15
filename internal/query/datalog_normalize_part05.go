package query

func datalogTermCanonical(term DatalogTerm) string {
	if term.Variable != "" {
		return "variable\x00" + term.Variable
	}
	return "constant\x00" + term.Constant.String()
}
