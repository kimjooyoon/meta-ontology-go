package workfrontier

func reverseRules(values []R4Rule) []R4Rule {
	result := append([]R4Rule(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
