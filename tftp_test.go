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
		out, err := fromNetASCII(tt.in)
		if err != nil {
			t.Fatal(err)
		}

		if want, got := tt.out, out; !bytes.Equal(want, got) {
			t.Fatalf("[%02d] unexpected fromNetASCII conversion:\n- want: %v\n-  got: %v",
				i, want, got)
		}
	}
}

// Test_toNetASCII verifies that toNetASCII properly converts a byte slice
// to netascii transfer mode form.
func Test_toNetASCII(t *testing.T) {
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
		if want, got := tt.out, toNetASCII(tt.in); !bytes.Equal(want, got) {
			t.Fatalf("[%02d] unexpected toNetASCII conversion:\n- want: %v\n-  got: %v",
				i, want, got)
		}
	}
}
