package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
)

// NamedInputDeclarationColumns retains authored names and their own byte ranges.
// Empty names remain empty; canonical ports carry the positional binding.
type NamedInputDeclarationColumns struct {
	InputDeclarationColumns
	Names      []string `json:"names"`
	NameStarts []int    `json:"name_starts"`
	NameEnds   []int    `json:"name_ends"`
}

func TypedNamedInputDeclarationInputs(r TypedValueOrigins, source []byte, target string) ([]byte, error) {
	if digest(source) != r.SourceSHA256 {
		return nil, fmt.Errorf("named input declarations differ from source digest")
	}
	rows, count, err := inputDeclarationRows(r)
	if err != nil {
		return nil, err
	}
	base, err := inputColumnsFromRows(r, rows, count)
	if err != nil {
		return nil, err
	}
	base.Schema = "gooo/input-declaration-columns/v2"
	c := NamedInputDeclarationColumns{InputDeclarationColumns: base,
		Names: make([]string, count), NameStarts: make([]int, count), NameEnds: make([]int, count)}
	previousEnd := 0
	for i := range count {
		d := rows[i]
		if d.Start < previousEnd || d.End > len(source) || string(source[d.Start:d.End]) != d.TypeName ||
			d.TypeID == "" || d.Port != inputbinding.PositionName(count, i) {
			return nil, fmt.Errorf("named input signature differs from original source position")
		}
		if err := validateInputNameSource(*d, source); err != nil {
			return nil, err
		}
		if d.Name != "" && d.NameStart < previousEnd {
			return nil, fmt.Errorf("named input range overlaps preceding input")
		}
		c.Names[i], c.NameStarts[i], c.NameEnds[i] = d.Name, d.NameStart, d.NameEnd
		previousEnd = d.End
	}
	return marshalDeclarationInputs(target, c)
}
