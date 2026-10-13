package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

const constructionObservationLimit = 4096
const constructionElementLimit = 65536

type CompositionConstructionElementSite struct {
	Expression string                            `json:"expression"`
	Start      int                               `json:"start_byte"`
	End        int                               `json:"end_byte"`
	Source     *bodycodegen.BodyExpressionSource `json:"source"`
}

type CompositionConstructionSite struct {
	ExpressionID     string                               `json:"expression_id"`
	RootActivityID   string                               `json:"root_activity_id"`
	ActivityID       string                               `json:"activity_id"`
	NativeOrdinal    int                                  `json:"native_ordinal"`
	ElementType      string                               `json:"element_type"`
	Expression       string                               `json:"expression"`
	ProjectionSHA256 string                               `json:"projection_sha256"`
	Start            int                                  `json:"start_byte"`
	End              int                                  `json:"end_byte"`
	Source           *bodycodegen.BodyExpressionSource    `json:"source"`
	Elements         []CompositionConstructionElementSite `json:"elements"`
}

type CompositionConstructedOrigin struct {
	Instance int `json:"instance_index"`
	Index    int `json:"element_index"`
}

// ReadStart and ReadEnd refer to the same case's Reads, as a half-open range.
// These are reads during evaluation, not a scalar data-dependency claim.
type CompositionConstructionElement struct {
	ReadStart    int  `json:"read_start"`
	ReadEnd      int  `json:"read_end"`
	ReadsDropped int  `json:"reads_dropped"`
	Completed    bool `json:"completed"`
}

type CompositionConstruction struct {
	SiteIndex int                              `json:"site_index"`
	ReadStart int                              `json:"read_start"`
	ReadEnd   int                              `json:"read_end"`
	Completed bool                             `json:"completed"`
	Elements  []CompositionConstructionElement `json:"elements"`
}

type compositionConstructionContext struct {
	sites   []CompositionConstructionSite
	rows    [][]CompositionConstruction
	dropped []int
}
