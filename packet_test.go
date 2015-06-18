package tftp

import (
	"reflect"
	"testing"
)

// Test_parseRequestPacket verifies that parseRequestPacket returns a correct
// requestPacket or error for an input byte slice.
func Test_parseRequestPacket(t *testing.T) {
	var tests = []struct {
		description string
		buf         []byte
		rp          *requestPacket
		err         error
	}{
		{
			description: "nil buffer, invalid request packet",
			err:         errInvalidRequestPacket,
		},
		{
			description: "length 3 buffer, invalid request packet",
			buf:         []byte{0, 0, 0},
			err:         errInvalidRequestPacket,
		},
		{
			description: "invalid opcode, invalid request packet",
			buf:         []byte{0, 3, 0, 0},
			err:         errInvalidRequestPacket,
		},
		{
			description: "opcode, filename, no trailing NULL, invalid request packet",
			buf:         []byte{0, 1, 'a', 255},
			err:         errInvalidRequestPacket,
		},
		{
			description: "opcode, filename, octet mode, no trailing NULL, invalid request packet",
			buf:         []byte{0, 1, 'a', 0, 'o', 'c', 't', 'e', 't'},
			err:         errInvalidRequestPacket,
		},
		{
			description: "opcode, filename, invalid mode, invalid request packet",
			buf:         []byte{0, 1, 'a', 0, 'o', 'c', 't', 'e', 'x', 0},
			err:         errInvalidRequestPacket,
		},
		{
			description: "opcode, filename, netascii mode, last byte not NULL, invalid request packet",
			buf:         []byte{0, 1, 'a', 0, 'N', 'e', 't', 'A', 'S', 'C', 'I', 'I', 0, 255},
			err:         errInvalidRequestPacket,
		},
		{
			description: "opcode, filename, netascii mode, OK",
			buf:         []byte{0, 1, 'a', 0, 'N', 'e', 't', 'A', 'S', 'C', 'I', 'I', 0},
			rp: &requestPacket{
				Opcode:   OpcodeRead,
				Filename: "a",
				Mode:     ModeNetASCII,
			},
		},
		{
			description: "opcode, filename, octet mode, OK",
			buf:         []byte{0, 2, 'b', 0, 'O', 'c', 'T', 'e', 'T', 0},
			rp: &requestPacket{
				Opcode:   OpcodeWrite,
				Filename: "b",
				Mode:     ModeOctet,
			},
		},
	}

	for i, tt := range tests {
		rp, err := parseRequestPacket(tt.buf)
		if err != nil {
			if want, got := tt.err, err; want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.description, want, got)
			}

			continue
		}

		if want, got := tt.rp, rp; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected packet:\n- want: %v\n-  got: %v",
				i, tt.description, want, got)
		}
	}
}

// Test_parseACKPacket verifies that parseACKPacket returns a correct
// ackPacket or error (possibly an ErrorPacket) for an input byte slice.
func Test_parseACKPacket(t *testing.T) {
	var tests = []struct {
		description string
		buf         []byte
		ack         *ackPacket
		err         error
	}{
		{
			description: "nil buffer, invalid ACK packet",
			err:         errInvalidACKPacket,
		},
		{
			description: "length 3 buffer, invalid ACK packet",
			buf:         []byte{0, 0, 0},
			err:         errInvalidACKPacket,
		},
		{
			description: "ACK packet, block 1, OK",
			buf:         []byte{0, 4, 0, 1},
			ack: &ackPacket{
				Opcode: opcodeACK,
				Block:  1,
			},
		},
		{
			description: "wrong opcode, invalid ERROR packet",
			buf:         []byte{0, 1, 0, 0},
			err:         errInvalidERRORPacket,
		},
		{
			description: "length 4 buffer, invalid ERROR packet",
			buf:         []byte{0, 5, 0, 0},
			err:         errInvalidERRORPacket,
		},
		{
			description: "ERROR packet, undefined code, empty message, no trailing NULL, invalid ERROR packet",
			buf:         []byte{0, 5, 0, 0, 255},
			err:         errInvalidERRORPacket,
		},
		{
			description: "ERROR packet, file not found, no message, OK",
			buf:         []byte{0, 5, 0, 1, 0},
			err: &ErrorPacket{
				Opcode:    OpcodeError,
				ErrorCode: ErrorCodeFileNotFound,
			},
		},
		{
			description: "ERROR packet, disk full, 'abc' message, OK",
			buf:         []byte{0, 5, 0, 3, 'a', 'b', 'c', 0},
			err: &ErrorPacket{
				Opcode:    OpcodeError,
				ErrorCode: ErrorCodeDiskFull,
				ErrorMsg:  "abc",
			},
		},
	}

	for i, tt := range tests {
		ack, err := parseACKPacket(tt.buf)
		if err != nil {
			if want, got := tt.err.Error(), err.Error(); want != got {
				t.Fatalf("[%02d] test %q, unexpected error: %v != %v",
					i, tt.description, want, got)
			}

			continue
		}

		if want, got := tt.ack, ack; !reflect.DeepEqual(want, got) {
			t.Fatalf("[%02d] test %q, unexpected packet:\n- want: %v\n-  got: %v",
				i, tt.description, want, got)
		}
	}
}
