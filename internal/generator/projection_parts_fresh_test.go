package generator

import (
	"bytes"
	"strings"
	"testing"
)

const ownedProjectionFixture = `package fixture

//gooo:generated:start id="fixture://record" kind="entity"
type Item struct {
    Value int64
}
func value() Item {
    return Item{Value: 1}
}
//gooo:generated:end id="fixture://record" kind="entity"

//gooo:generated:start id="fixture://activity" kind="activity"
func choose() int64 {
    return value().Value
}
//gooo:generated:end id="fixture://activity" kind="activity"
`

func TestOwnedProjectionPartitionFresh1657(t *testing.T) {
	parts, err := PartitionProjection([]byte(ownedProjectionFixture), 14)
	if err != nil || len(parts) != 2 {
		t.Fatal("owned region partition", len(parts), err)
	}
	for i, part := range parts {
		if bytes.Count(part.Source, []byte{'\n'}) > 14 {
			t.Fatal("physical file budget exceeded")
		}
		markers, err := parseMarkers(part.Source)
		if err != nil || len(markers.Regions) != 1 || len(part.RegionIDs) != 1 {
			t.Fatal("a generated region was fragmented", err)
		}
		if part.RegionIDs[0] != []string{"fixture://record", "fixture://activity"}[i] {
			t.Fatal("region identity or source order changed")
		}
	}
}

func TestOwnedProjectionRefusalsFresh1657(t *testing.T) {
	for _, item := range []struct {
		name, source string
		budget       int
	}{
		{"indivisible-region", ownedProjectionFixture, 4},
		{"missing-end", strings.Replace(ownedProjectionFixture,
			`//gooo:generated:end id="fixture://activity" kind="activity"`, "", 1), 75},
		{"unowned-code", ownedProjectionFixture + "var outside = 1\n", 75},
		{"import-preamble", strings.Replace(ownedProjectionFixture, "package fixture", "package fixture\nimport \"fmt\"", 1), 75},
		{"empty-ownership", "package fixture\n", 75},
	} {
		t.Run(item.name, func(t *testing.T) {
			parts, err := PartitionProjection([]byte(item.source), item.budget)
			if err == nil || parts != nil {
				t.Fatal("invalid partition returned partial output")
			}
		})
	}
}
