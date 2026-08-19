package bidir

import (
	"sort"
	"strings"
)

func joinIDs(ids []ID) string {
	copyIDs := append([]ID(nil), ids...)
	sort.Slice(copyIDs, func(i, j int) bool { return copyIDs[i] < copyIDs[j] })
	values := make([]string, len(copyIDs))
	for index, id := range copyIDs {
		values[index] = string(id)
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
