package proofchoicealgebra

import "fmt"

type issue struct {
	Reason string
	Line   int
}

func (i issue) Error() string {
	if i.Line > 0 {
		return fmt.Sprintf("%s at line %d", i.Reason, i.Line)
	}
	return i.Reason
}
