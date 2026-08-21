package generator

import (
	"bytes"
)

func patchExisting(previous []byte, regions []generatedRegion, blocks map[string][]byte, order []string) []byte {
	var output bytes.Buffer
	cursor := 0
	present := make(map[string]struct{}, len(regions))
	for _, region := range regions {
		output.Write(previous[cursor:region.Start])
		if block, exists := blocks[region.ID]; exists {
			output.Write(block)
			present[region.ID] = struct{}{}
		}
		cursor = region.End
	}
	output.Write(previous[cursor:])

	for _, id := range order {
		if _, exists := present[id]; exists {
			continue
		}
		appendGeneratedBlock(&output, blocks[id])
	}
	return output.Bytes()
}
func appendGeneratedBlock(output *bytes.Buffer, block []byte) {
	value := output.Bytes()
	if len(value) > 0 && value[len(value)-1] != '\n' {
		output.WriteByte('\n')
	}
	if output.Len() > 0 {
		output.WriteByte('\n')
	}
	output.Write(block)
}
