package cycles

import (
	"sort"
)

func cyclePath(component []ID, adjacency map[ID][]ID) []ID {
	start := component[0]
	members := make(map[ID]struct{}, len(component))
	for _, id := range component {
		members[id] = struct{}{}
	}

	reverse := make(map[ID][]ID, len(component))
	for _, id := range component {
		for _, next := range adjacency[id] {
			if _, member := members[next]; member {
				reverse[next] = append(reverse[next], id)
			}
		}
	}
	for id := range reverse {
		reverse[id] = uniqueSorted(reverse[id])
	}
	queue := []ID{start}
	nextTowardStart := map[ID]ID{start: ""}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, predecessor := range reverse[current] {
			if _, seen := nextTowardStart[predecessor]; seen {
				continue
			}
			nextTowardStart[predecessor] = current
			queue = append(queue, predecessor)
		}
	}

	for _, next := range adjacency[start] {
		if _, member := members[next]; !member {
			continue
		}
		if next == start {
			return []ID{start, start}
		}
		if _, reachable := nextTowardStart[next]; !reachable {
			continue
		}
		path := []ID{start, next}
		for current := nextTowardStart[next]; current != start; current = nextTowardStart[current] {
			path = append(path, current)
		}
		return append(path, start)
	}
	return []ID{start, start}
}
func uniqueSorted(values []ID) []ID {
	result := append([]ID(nil), values...)
	sort.Strings(result)
	write := 0
	for _, value := range result {
		if write == 0 || result[write-1] != value {
			result[write] = value
			write++
		}
	}
	return result[:write]
}
