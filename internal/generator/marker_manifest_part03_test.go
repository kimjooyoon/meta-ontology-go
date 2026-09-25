package generator

func markerRegionIndexV1(markers parsedMarkers, id string) int {
	for index, region := range markers.Regions {
		if region.ID == id {
			return index
		}
	}
	return -1
}
func cloneParsedMarkersV1(markers parsedMarkers) parsedMarkers {
	clone := parsedMarkers{Regions: make([]generatedRegion, len(markers.Regions)), Slots: make(map[string]parsedSlot, len(markers.Slots))}
	for index, region := range markers.Regions {
		clone.Regions[index] = region
		clone.Regions[index].Slots = append([]parsedSlot(nil), region.Slots...)
		for slotIndex := range clone.Regions[index].Slots {
			clone.Regions[index].Slots[slotIndex].Body = append([]byte(nil), region.Slots[slotIndex].Body...)
		}
	}
	for id, slot := range markers.Slots {
		slot.Body = append([]byte(nil), slot.Body...)
		clone.Slots[id] = slot
	}
	return clone
}
