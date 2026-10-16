package generator

import "fmt"

// The tag describes the source constraint for Go readers. Gooo's typed runtime
// owns value admission; a Go struct tag does not execute a validator by itself.
func fieldTextBoundTag(field Field) string {
	if !field.TextBound.IsSet() {
		return ""
	}
	maximum, _ := field.TextBound.Resolve(0)
	return fmt.Sprintf(" `gooo_max_utf8_bytes:%q`", fmt.Sprint(maximum))
}
