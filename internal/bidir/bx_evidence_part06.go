package bidir

import (
	"slices"
	"strings"
)

func joinIDs(ids []ID) string {
	copyIDs := append([]ID(nil), ids...)
	slices.Sort(copyIDs)
	values := make([]string, len(copyIDs))
	for index, id := range copyIDs {
		values[index] = string(id)
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
