package textbound

import (
	"bytes"
	"strconv"
)

func (b UTF8) MarshalJSON() ([]byte, error) {
	maximum, err := b.Resolve(0)
	if err != nil {
		return nil, err
	}
	if !b.IsSet() {
		return []byte("null"), nil
	}
	return strconv.AppendUint(nil, uint64(maximum), 10), nil
}

func (b *UTF8) UnmarshalJSON(raw []byte) error {
	if b == nil {
		return ErrInvalid
	}
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		*b = 0
		return nil
	}
	parsed, err := Parse(string(raw))
	if err != nil {
		return err
	}
	*b = parsed
	return nil
}
