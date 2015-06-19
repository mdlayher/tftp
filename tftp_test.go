package tftp

import (
	"bytes"
	"testing"
)

// Test_fromNetASCII verifies that fromNetASCII properly converts a byte slice
// from netascii transfer mode form.
func Test_fromNetASCII(t *testing.T) {
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
			in:  []byte{'a', '\r', '\n', 'b', '\r', '\n', 'c'},
			out: []byte{'a', '\n', 'b', '\n', 'c'},
		},
		{
			in:  []byte{'a', '\r', 0, 'b', '\r', 0, 'c'},
			out: []byte{'a', '\r', 'b', '\r', 'c'},
		},
		{
			in:  []byte{'a', '\r', 0, 'b', '\r', '\n', 'c'},
			out: []byte{'a', '\r', 'b', '\n', 'c'},
		},
		// TODO(mdlayher): determine if it possible for a carriage return to
		// be the last character in a buffer.  For the time being, we perform
		// no conversion if this is the case.
		{
			in:  []byte{'a', '\r'},
			out: []byte{'a', '\r'},
		},
	}

	for i, tt := range tests {
		if want, got := tt.out, fromNetASCII(tt.in); !bytes.Equal(want, got) {
			t.Fatalf("[%02d] unexpected fromNetASCII conversion:\n- want: %v\n-  got: %v",
				i, want, got)
		}
	}
}
