package bodycodegen

import "fmt"

// Each filled expression uses the same source-owned call closure as ordinary
// body generation. The unfilled hole is not an executable base program.
func prepareSearchBody(filename string, source []byte, activity, hole string) (preparedBody, error) {
	p, err := prepareActivityBodySource(filename, source, activity)
	if err != nil {
		return p, err
	}
	if len(p.parameters) != 1 || p.parameters[0].Type != "int64" || p.outputType != "int64" {
		return p, fmt.Errorf("IR body search requires one Integer input and one Integer output")
	}
	if countIdentifier(p.body, bodyFillHoleToken(hole)) != 1 {
		return p, fmt.Errorf("activity %q must contain exactly one body-search hole", activity)
	}
	return p, nil
}
