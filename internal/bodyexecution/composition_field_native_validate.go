package bodyexecution

import (
	"cmp"
	"fmt"
	"slices"
)

type nativeFieldFile struct {
	Schema       string  `json:"schema"`
	Writes       [][]int `json:"writes"`
	Uses         [][]int `json:"uses"`
	Frames       []int   `json:"frames"`
	FrameDropped []int   `json:"frame_dropped"`
	WriteDropped []int   `json:"write_dropped"`
	UseDropped   []int   `json:"use_dropped"`
}

// Validate recorded relationships and temporal ordering, independently of values.
// This checks a recorded claim; it does not authenticate execution of a producer.
func validateNativeFieldFile(f nativeFieldFile, p *compositionFieldProjection, cases int) error {
	if err := validateFieldInventory(p); err != nil {
		return err
	}
	if f.Schema != "gooo/native-field-observation/v1" || cases < 1 || cases > 128 ||
		f.Writes == nil || f.Uses == nil || len(f.Writes) > 4096 || len(f.Uses) > 65536 ||
		len(f.Frames) != cases || len(f.FrameDropped) != cases || len(f.WriteDropped) != cases || len(f.UseDropped) != cases {
		return fmt.Errorf("native field protocol or bounds differ")
	}
	groups := [2][][]int{f.Writes, f.Uses}
	widths := [2]int{9, 5}
	for group, rows := range groups {
		last := -1
		for _, row := range rows {
			limit := len(p.sites)
			if group == 1 {
				limit = len(p.uses)
			}
			if len(row) != widths[group] || row[0] < 0 || row[0] >= cases || row[0] < last || row[1] < 0 || row[1] >= limit || row[2] < 0 || row[2] >= f.Frames[row[0]] {
				return fmt.Errorf("native field row shape, site, case or frame differs")
			}
			last = row[0]
		}
	}
	frames := 0
	for c := range cases {
		if f.Frames[c] < 0 || f.Frames[c] > 65536 || f.FrameDropped[c] < 0 || f.WriteDropped[c] < 0 || f.UseDropped[c] < 0 {
			return fmt.Errorf("native field counters are negative")
		}
		frames += f.Frames[c]
		if frames > 65536 {
			return fmt.Errorf("field frame total exceeds65536")
		}
		var writes, uses [][]int
		for _, row := range f.Writes {
			if row[0] == c {
				writes = append(writes, row)
			}
		}
		for _, row := range f.Uses {
			if row[0] == c {
				uses = append(uses, row)
			}
		}
		if err := validateNativeFieldCase(writes, uses, p); err != nil {
			return fmt.Errorf("field case%d: %w", c, err)
		}
	}
	return nil
}

func validateNativeFieldCase(writes, uses [][]int, p *compositionFieldProjection) error {
	var completed []int
	for i, row := range writes {
		if row[3] < 0 || row[4] < row[3] || row[4] > len(uses) || row[5] < 0 || (row[6] != 0 && row[6] != 1) || row[7] < -2 || row[7] >= len(writes) {
			return fmt.Errorf("field write window, completion or parent differs")
		}
		if row[6] == 1 {
			completed = append(completed, i)
		} else if row[8] != -1 {
			return fmt.Errorf("incomplete field write has completion order")
		}
		if row[7] >= 0 {
			parent := writes[row[7]]
			a, b := p.sites[row[1]], p.sites[parent[1]]
			x, y := p.inventory.bindings[a.Binding], p.inventory.bindings[b.Binding]
			if parent[6] != 1 || row[6] == 1 && parent[8] >= row[8] || x.Record.ID != y.Record.ID || x.Record.Fields[a.Field].ID != y.Record.Fields[b.Field].ID {
				return fmt.Errorf("field copy/argument/return parent identity or completion differs")
			}
		}
	}
	slices.SortFunc(completed, func(a, b int) int { return cmp.Compare(writes[a][8], writes[b][8]) })
	for i, index := range completed {
		if writes[index][8] != i {
			return fmt.Errorf("field completion order is not distinct and contiguous")
		}
	}
	type key struct{ frame, binding, field int }
	latest := map[key]int{}
	cursor, previousClock := 0, 0
	for _, use := range uses {
		if use[3] < -2 || use[3] >= len(writes) || use[4] < previousClock || use[4] > len(completed) {
			return fmt.Errorf("field read assignment or completion clock differs")
		}
		previousClock = use[4]
		for cursor < len(completed) && writes[completed[cursor]][8] < use[4] {
			index := completed[cursor]
			write := writes[index]
			site := p.sites[write[1]]
			latest[key{write[2], site.Binding, site.Field}] = index
			cursor++
		}
		site := p.uses[use[1]]
		want, found := latest[key{use[2], site.Binding, site.Field}]
		if use[3] >= 0 && (!found || use[3] != want) {
			return fmt.Errorf("field read does not reference its latest completed field write")
		}
		if use[3] == -1 && found {
			return fmt.Errorf("field read discarded an available completed assignment")
		}
	}
	return nil
}
