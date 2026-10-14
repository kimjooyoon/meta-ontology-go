package bodycodegen

// BodyFieldSource maps an authored value and its field target into the selected
// Gooo literal. A later workspace pass can undo package-name lowering.
type BodyFieldSource struct {
	Value  *BodyExpressionSource `json:"value"`
	Target *BodyExpressionSource `json:"target"`
}

func BindBodyFieldSources(source []byte, origins []BodyFieldOrigin) ([]BodyFieldSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	sha := digest(source)
	result := make([]BodyFieldSource, len(origins))
	for i, origin := range origins {
		activity := activities[origin.ActivityID]
		result[i] = BodyFieldSource{bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, origin.Value), bindBodyExpressionSource(source, sha, activity, origin.BodySHA256, origin.Target)}
	}
	return result, nil
}
