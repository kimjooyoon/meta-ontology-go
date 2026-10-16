package executioncontinuation

import (
	"context"
	"encoding/base64"
	"strings"

	input "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

func (c Client) sourceBytes(ctx context.Context, subject, path string) ([]byte, string, error) {
	if !validSHA(subject) || path != input.SourcePath {
		return nil, "", refuted("SOURCE_PATH_REFUTED")
	}
	var file struct {
		Content, Encoding, SHA string
		Size                   int64
	}
	if err := c.get(ctx, "contents/"+path+"?ref="+subject, &file); err != nil {
		return nil, "", unknown("SUBJECT_SOURCE_UNKNOWN")
	}
	if file.Encoding != "base64" || file.Size <= 0 || file.Size > 8<<20 || !validSHA(file.SHA) {
		return nil, "", refuted("SUBJECT_SOURCE_INVALID")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
	if err != nil || int64(len(raw)) != file.Size {
		return nil, "", refuted("SUBJECT_SOURCE_INVALID")
	}
	return raw, file.SHA, nil
}
