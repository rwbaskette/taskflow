package cmd

import (
	"bytes"
)

// testBuffer implements io.Writer for capturing command output in tests
type testBuffer struct {
	buf *bytes.Buffer
}

func (t *testBuffer) Write(p []byte) (n int, err error) {
	return t.buf.Write(p)
}

func (t *testBuffer) String() string {
	return t.buf.String()
}

// Stringer is an interface that provides String() method
type Stringer interface {
	String() string
}

// newTestBuffer creates a buffer for capturing command output in tests
func newTestBuffer() Stringer {
	return &testBuffer{buf: &bytes.Buffer{}}
}

// OutputBuffer creates a buffer that implements both io.Writer and Stringer
type OutputBuffer struct {
	*bytes.Buffer
}

// NewOutputBuffer creates a new OutputBuffer
func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{Buffer: &bytes.Buffer{}}
}
