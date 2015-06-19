package tftp

import (
	"bytes"
)

// netASCIIResponseWriter is a ResponseWriter which seamlessly writes input
// data in netascii format to the embedded ResponseWriter.
type netASCIIResponseWriter struct {
	ResponseWriter
}

// Write converts data to netascii format and writes it to the embedded
// ResponseWriter.
func (w *netASCIIResponseWriter) Write(p []byte) (int, error) {
	b := make([]byte, len(p))
	copy(b, p)

	// If using netascii mode, some conversions must be made to
	// input data:
	//   - LF -> CR+LF
	//   - CR -> CR+NULL
	b = bytes.Replace(b, []byte{'\r'}, []byte{'\r', 0}, -1)
	b = bytes.Replace(b, []byte{'\n'}, []byte{'\r', '\n'}, -1)

	_, err := w.ResponseWriter.Write(b)
	return len(p), err
}
