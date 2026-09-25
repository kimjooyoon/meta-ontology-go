package bidir

import (
	"sort"
)

// Normalize sorts delta members and recalculates relation IDs.
func (d *Delta) Normalize() {
	if d == nil {
		return
	}
	for index := range d.AddedNodes {
		d.AddedNodes[index] = d.AddedNodes[index].normalized()
	}
	for index := range d.RemovedNodes {
		d.RemovedNodes[index] = d.RemovedNodes[index].normalized()
	}
	for index := range d.AddedRelations {
		d.AddedRelations[index] = d.AddedRelations[index].normalized()
	}
	for index := range d.RemovedRelations {
		d.RemovedRelations[index] = d.RemovedRelations[index].normalized()
	}
	sort.Slice(d.AddedNodes, func(i, j int) bool { return d.AddedNodes[i].ID < d.AddedNodes[j].ID })
	sort.Slice(d.RemovedNodes, func(i, j int) bool { return d.RemovedNodes[i].ID < d.RemovedNodes[j].ID })
	sort.Slice(d.AddedRelations, func(i, j int) bool { return relationLess(d.AddedRelations[i], d.AddedRelations[j]) })
	sort.Slice(d.RemovedRelations, func(i, j int) bool { return relationLess(d.RemovedRelations[i], d.RemovedRelations[j]) })
}

// IsEmpty reports whether the delta changes semantic content.
func (d Delta) IsEmpty() bool {
	return len(d.AddedNodes) == 0 && len(d.RemovedNodes) == 0 && len(d.AddedRelations) == 0 && len(d.RemovedRelations) == 0
}
