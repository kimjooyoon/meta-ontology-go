package coupling

import (
	"sort"
)

func normalizePathWire(path *wirePath) {
	sort.Slice(path.Edges, func(i, j int) bool { return path.Edges[i].RecordID < path.Edges[j].RecordID })
	sort.Slice(path.Claims, func(i, j int) bool { return path.Claims[i].RecordID < path.Claims[j].RecordID })
	sort.Slice(path.Evidence, func(i, j int) bool { return path.Evidence[i].ID < path.Evidence[j].ID })
	for i := range path.Edges {
		sort.Strings(path.Edges[i].SourceRoots)
		sort.Slice(path.Edges[i].Evidence, func(a, b int) bool { return path.Edges[i].Evidence[a].ID < path.Edges[i].Evidence[b].ID })
	}
	for i := range path.Claims {
		sort.Slice(path.Claims[i].Evidence, func(a, b int) bool { return path.Claims[i].Evidence[a].ID < path.Claims[i].Evidence[b].ID })
	}
}
