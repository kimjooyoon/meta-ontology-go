package bodycodegen

type BodyLocalSource struct {
	Value       *BodyExpressionSource `json:"value"`
	Declaration *BodyExpressionSource `json:"declaration"`
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
			Value:       bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, origin.BodyElementOrigin),
			Declaration: bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, origin.Declaration),
		}
	}
	return result, nil
}
