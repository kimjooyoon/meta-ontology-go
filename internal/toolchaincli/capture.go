package toolchaincli

const maxOutputBytes = 64 * 1024

type cappedBuffer struct {
	data     []byte
	overflow bool
}

func (buffer *cappedBuffer) Write(value []byte) (int, error) {
	written := len(value)
	remaining := maxOutputBytes - len(buffer.data)
	if remaining <= 0 {
		buffer.overflow = true
		return written, nil
	}
	if len(value) > remaining {
		buffer.data = append(buffer.data, value[:remaining]...)
		buffer.overflow = true
		return written, nil
	}
	buffer.data = append(buffer.data, value...)
	return written, nil
}

func (buffer *cappedBuffer) String() string { return string(buffer.data) }
