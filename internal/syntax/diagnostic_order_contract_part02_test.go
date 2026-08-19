package syntax

func diagnosticLabels(diagnostics Diagnostics) []string {
	labels := make([]string, len(diagnostics))
	for index, diagnostic := range diagnostics {
		labels[index] = diagnostic.Span.Filename + ":" + diagnostic.Message
	}
	return labels
}
func diagnosticPermutations(input Diagnostics) []Diagnostics {
	result := make([]Diagnostics, 0)
	var visit func(int)
	current := append(Diagnostics(nil), input...)
	visit = func(index int) {
		if index == len(current) {
			result = append(result, append(Diagnostics(nil), current...))
			return
		}
		for next := index; next < len(current); next++ {
			current[index], current[next] = current[next], current[index]
			visit(index + 1)
			current[index], current[next] = current[next], current[index]
		}
	}
	visit(0)
	return result
}
