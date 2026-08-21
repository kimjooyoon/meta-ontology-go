package shadow

import (
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
)

type productionFixture struct {
	files      map[string][]byte
	base       analyzersci.Snapshot
	head       analyzersci.Snapshot
	planInput  plannersci.Input
	proofInput proofsci.Input
	laneInput  lanesci.Input
	entityID   string
	otherID    string
	commandID  string
	sourceBase string
}
