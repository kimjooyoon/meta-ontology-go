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
	Kind             string                               `json:"kind,omitempty"`
	MapperActivityID string                               `json:"mapper_activity_id,omitempty"`
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
	SiteIndex    int                              `json:"site_index"`
	ReadStart    int                              `json:"read_start"`
	ReadEnd      int                              `json:"read_end"`
	Completed    bool                             `json:"completed"`
	Elements     []CompositionConstructionElement `json:"elements"`
	ElementCount *int                             `json:"element_count,omitempty"`
}

func constructionElementSite(site CompositionConstructionSite, index int) (CompositionConstructionElementSite, bool) {
	if site.Kind == "map_values" && index >= 0 && index < 256 && len(site.Elements) == 1 {
		return site.Elements[0], true
	}
	if site.Kind == "" && index >= 0 && index < len(site.Elements) {
		return site.Elements[index], true
	}
	return CompositionConstructionElementSite{}, false
}

func constructionSize(site CompositionConstructionSite, count *int) (int, bool) {
	if site.Kind == "" {
		return len(site.Elements), count == nil && len(site.Elements) <= 256
	}
	if site.Kind == "map_values" && count != nil && *count >= 0 && *count <= 256 && len(site.Elements) == 1 && site.MapperActivityID != "" {
		return *count, true
	}
	return 0, false
}

type compositionConstructionContext struct {
	sites   []CompositionConstructionSite
	rows    [][]CompositionConstruction
	dropped []int
}
