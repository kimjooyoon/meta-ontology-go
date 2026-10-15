package bodycodegen

import "fmt"

func BindBodyReturnSources(source []byte, origins []BodyReturnOrigin) ([]*BodyExpressionSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	sha := digest(source)
	result := make([]*BodyExpressionSource, len(origins))
	for i, origin := range origins {
		if origin.ActivityID == "" || origin.RecordID == "" || origin.NativeOrdinal < 0 || origin.NativeOrdinal >= origin.NativePopulation {
			return nil, fmt.Errorf("return source typed identity or population differs")
		}
		if origin.Value == nil {
			if origin.Reason == "" || origin.BodySHA256 != "" {
				return nil, fmt.Errorf("unretained return claims authored source")
			}
			result[i] = UnmappedReadSource(sha, origin.Reason)
			continue
		}
		if origin.Reason != "" || origin.BodySHA256 == "" {
			return nil, fmt.Errorf("retained return source is incomplete")
		}
		result[i] = bindBodyExpressionSource(source, sha, activities[origin.ActivityID], origin.BodySHA256, *origin.Value)
	}
	return result, nil
}
