package bidir

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func cloneLowerDocument(document Document) Document {
	clone := document
	clone.Declarations = append([]Declaration(nil), document.Declarations...)
	for index := range clone.Declarations {
		clone.Declarations[index].Inputs = append([]Reference(nil), document.Declarations[index].Inputs...)
		clone.Declarations[index].Outputs = append([]Reference(nil), document.Declarations[index].Outputs...)
	}
	clone.Relations = append([]Relation(nil), document.Relations...)
	return clone
}
func zeroPad(value int) string {
	if value < 10 {
		return "00" + string(rune('0'+value))
	}
	if value < 100 {
		return "0" + string(rune('0'+value/10)) + string(rune('0'+value%10))
	}
	return string(rune('0'+value/100)) + string(rune('0'+(value/10)%10)) + string(rune('0'+value%10))
}

type cancelAfterContext struct {
	done   chan struct{}
	limit  int32
	checks atomic.Int32
	once   sync.Once
}

func newCancelAfterContext(limit int) *cancelAfterContext {
	return &cancelAfterContext{done: make(chan struct{}), limit: int32(limit)}
}
func (c *cancelAfterContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelAfterContext) Done() <-chan struct{} {
	if c.checks.Add(1) >= c.limit {
		c.once.Do(func() { close(c.done) })
	}
	return c.done
}
func (c *cancelAfterContext) Err() error {
	select {
	case <-c.done:
		return context.Canceled
	default:
		return nil
	}
}
func (c *cancelAfterContext) Value(any) any { return nil }
