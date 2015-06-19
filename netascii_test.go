package tftp

import (
	"bytes"
	"testing"
)

// Test_netASCIIResponseWriterWrite verifies that netASCIIResponseWriter.Write
// properly converts a byte slice to netascii transfer mode form.
func Test_netASCIIResponseWriterWrite(t *testing.T) {
	var tests = []struct {
		in  []byte
		out []byte
	}{
		{
			in:  nil,
			out: nil,
		},
		{
			in:  []byte{'a', 'b', 'c'},
			out: []byte{'a', 'b', 'c'},
		},
		{
			in:  []byte{'a', '\n', 'b', '\n', 'c'},
			out: []byte{'a', '\r', '\n', 'b', '\r', '\n', 'c'},
		},
		{
			in:  []byte{'a', '\r', 'b', '\r', 'c'},
			out: []byte{'a', '\r', 0, 'b', '\r', 0, 'c'},
		},
		{
			in:  []byte{'a', '\r', 'b', '\n', 'c'},
			out: []byte{'a', '\r', 0, 'b', '\r', '\n', 'c'},
		},
	}

	for i, tt := range tests {
		b := bytes.NewBuffer(nil)
		w := &netASCIIResponseWriter{&captureResponseWriter{
			buf: b,
		}}
		w.Write(tt.in)

		if want, got := tt.out, b.Bytes(); !bytes.Equal(want, got) {
			t.Fatalf("[%02d] unexpected toNetASCII conversion:\n- want: %v\n-  got: %v",
				i, want, got)
		}
	}
}

// captureResponseWriter captures any data written to it using a buffer.
type captureResponseWriter struct {
	buf *bytes.Buffer
}

func (w *captureResponseWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *captureResponseWriter) Close() error                { return nil }
func (w *captureResponseWriter) Flush() error                { return nil }
