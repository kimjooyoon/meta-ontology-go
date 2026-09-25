package languagesemantic

import (
	"fmt"
	"strings"
)

func validateHeadSHA(value string) error {
	if len(value) != 40 {
		return fmt.Errorf("head SHA must contain 40 hexadecimal characters")
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return fmt.Errorf("head SHA must be lowercase hexadecimal")
		}
	}
	return nil
}
