package bindingcoverage

func reversePrecedence(values []PrecedenceEntry) []PrecedenceEntry {
	result := append([]PrecedenceEntry{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
