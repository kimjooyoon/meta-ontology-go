package bodycodegen

type BodyLocalSource struct {
	Value       *BodyExpressionSource `json:"value"`
	Declaration *BodyExpressionSource `json:"declaration"`
	Uses        []BodyLocalUseSource  `json:"uses,omitempty"`
}

func BindBodyLocalSources(source []byte, origins []BodyLocalOrigin) ([]BodyLocalSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	sha := digest(source)
	result := make([]BodyLocalSource, len(origins))
	for i, origin := range origins {
		activity := activities[origin.ActivityID]
		result[i] = BodyLocalSource{
			Value:       bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, BodyElementOrigin{origin.Start, origin.End, origin.Expression}),
			Declaration: bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, origin.Declaration),
		}
		for _, use := range origin.Uses {
			result[i].Uses = append(result[i].Uses, BodyLocalUseSource{NativeOrdinal: use.NativeOrdinal, Source: bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, BodyElementOrigin{use.Start, use.End, use.Expression})})
		}
	}
	return result, nil
}

type BodyLocalUseSource struct {
	NativeOrdinal int                   `json:"native_ordinal"`
	Source        *BodyExpressionSource `json:"source"`
}
