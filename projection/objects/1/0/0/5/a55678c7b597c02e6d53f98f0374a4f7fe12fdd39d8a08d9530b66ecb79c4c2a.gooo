package lsp

func completionItem(items []CompletionItem, label string) (CompletionItem, bool) {
	for _, item := range items {
		if item.Label == label {
			return item, true
		}
	}
	return CompletionItem{}, false
}
func containsID(ids map[string]string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
