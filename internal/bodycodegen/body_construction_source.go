package bodycodegen

type BodyConstructionSource struct {
	Literal          *BodyExpressionSource   `json:"literal"`
	Elements         []*BodyExpressionSource `json:"elements"`
	Kind             string                  `json:"kind,omitempty"`
	MapperActivityID string                  `json:"mapper_activity_id,omitempty"`
}

func BindBodyConstructionSources(source []byte, origins []BodyConstructionOrigin) ([]BodyConstructionSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	sha := digest(source)
	result := make([]BodyConstructionSource, len(origins))
	for i, o := range origins {
		result[i].Kind, result[i].MapperActivityID = o.Kind, o.MapperActivityID
		a := activities[o.ActivityID]
		result[i].Literal = bindBodyExpressionSource(source, sha, a, o.BodySHA256, o.BodyElementOrigin)
		result[i].Elements = make([]*BodyExpressionSource, len(o.Elements))
		for j, e := range o.Elements {
			result[i].Elements[j] = bindBodyExpressionSource(source, sha, a, o.BodySHA256, e)
		}
	}
	return result, nil
}
