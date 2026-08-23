package toolchainlsp

import "github.com/kimjooyoon/meta-ontology-go/internal/lsp"

func observeUTF16() bool {
	source := "a😀b\r\nΩ"
	position, err := lsp.OffsetToPosition(source, len("a😀"))
	if err != nil || position.Character != 3 {
		return false
	}
	offset, err := lsp.PositionToOffset(source, position)
	return err == nil && offset == len("a😀")
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}
