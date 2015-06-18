package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

var (
	// errInvalidRequestPacket is returned when an invalid TFTP request is
	// received, allowing the server to reject the request.
	errInvalidRequestPacket = errors.New("invalid request packet")

	// errInvalidACKPacket is returned when an invalid TFTP ACK packet is received.
	errInvalidACKPacket = errors.New("invalid ACK packet")

	// errInvalidERRORPacket is returned when an invalid TFTP ERROR packet is
	// received.
	errInvalidERRORPacket = errors.New("invalid ERROR packet")
)

// ErrorPacket represents an ERROR packet, as defined in RFC 1350, Section 5.
// ErrorPacket implements the error interface, and may be returned during a
// read or write operation with a client.
type ErrorPacket struct {
	Opcode    Opcode
	ErrorCode ErrorCode
	ErrorMsg  string
}

// Error returns the string representation of an ErrorPacket.
func (e *ErrorPacket) Error() string {
	return fmt.Sprintf("%s (%02d): %s", e.ErrorCode.String(), e.ErrorCode, e.ErrorMsg)
}

// requestPacket represents a raw request to a TFTP server.  It is used to
// construct a Request for client consumption.
type requestPacket struct {
	Opcode   Opcode
	Filename string
	Mode     Mode
}

// parseRequestPacket attempts to parse a TFTP read or write request as a
// requestPacket.
func parseRequestPacket(b []byte) (*requestPacket, error) {
	// At a minimum, requests must contain a 2 byte opcode and
	// 2 NULL bytes
	if len(b) < 4 {
		return nil, errInvalidRequestPacket
	}

	opcode := Opcode(binary.BigEndian.Uint16(b[0:2]))

	// Only accept read and write requests to start transactions
	if opcode != OpcodeRead && opcode != OpcodeWrite {
		return nil, errInvalidRequestPacket
	}

	// Locate first NULL byte to determine filename offset
	idx := bytes.IndexByte(b[2:], 0)
	if idx == -1 {
		return nil, errInvalidRequestPacket
	}

	filename := string(b[2 : idx+2])

	// Locate second NULL byte to determine mode offset
	offset := idx + 3
	idx = bytes.IndexByte(b[offset:], 0)
	if idx == -1 {
		return nil, errInvalidRequestPacket
	}

	// Mode can be any combination of uppercase and lowercase, but for our
	// purposes, we just want to deal with lowercase letters
	mode := Mode(strings.ToLower(string(b[offset : offset+idx])))

	// Only accept netascii or octet modes
	if mode != ModeNetASCII && mode != ModeOctet {
		return nil, errInvalidRequestPacket
	}

	// BUG(mdlayher): implement options extension from RFC 2347.

	// Trailing NULL byte must be present to end packet
	if b[len(b)-1] != 0 {
		return nil, errInvalidRequestPacket
	}

	return &requestPacket{
		Opcode:   opcode,
		Filename: filename,
		Mode:     mode,
	}, nil
}

// ackPacket represents an ACK packet, as defined in RFC 1350, Section 5.
// An ACK packet is used to confirm acknowledgement of receipt of a DATA
// packet.
type ackPacket struct {
	Opcode Opcode
	Block  uint16
}

// parseACKPacket attempts to parse an ackPacket from a byte slice, but may
// also return an ErrorPacket as the error value, if an error occurs.
func parseACKPacket(b []byte) (*ackPacket, error) {
	// At a minimum, ACK packet must contain a 2 byte opcode and a 2 byte
	// block number
	if len(b) < 4 {
		return nil, errInvalidACKPacket
	}

	opcode := Opcode(binary.BigEndian.Uint16(b[0:2]))
	n := binary.BigEndian.Uint16(b[2:4])

	// If length is exactly 4 and opcode is correct, return an ACK packet
	if len(b) == 4 && opcode == opcodeACK {
		return &ackPacket{
			Opcode: opcode,
			Block:  n,
		}, nil
	}

	// Verify packet is an ERROR packet
	if opcode != OpcodeError {
		return nil, errInvalidERRORPacket
	}

	// At a minimum, ERROR packet must contain:
	//  - 2 bytes: opcode
	//  - 2 bytes: error code
	//  - 1 byte : NULL
	if len(b) < 5 {
		return nil, errInvalidERRORPacket
	}

	msg := string(b[4 : len(b)-1])

	// Trailing NULL byte mut be present to end packet
	if b[len(b)-1] != 0 {
		return nil, errInvalidERRORPacket
	}

	return nil, &ErrorPacket{
		Opcode:    opcode,
		ErrorCode: ErrorCode(n),
		ErrorMsg:  msg,
	}
}
