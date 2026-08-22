package directorykind

import (
	"fmt"
	"path"
)

func availableGroup(subject, base string, occupied map[string]bool) string {
	for index := 1; ; index++ {
		name := base
		if index > 1 {
			name = fmt.Sprintf("%s_%02d", base, index)
		}
		candidate := path.Join(subject, name)
		if !occupied[candidate] {
			occupied[candidate] = true
			return candidate
		}
	}
}
