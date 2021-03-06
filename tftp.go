// Package tftp implements a TFTP server, as described in RFC 1350.
package tftp

import (
	"bytes"
)

//go:generate stringer -output=string.go -type=ErrorCode,Opcode

// Opcode represents a TFTP opcode, as defined in RFC 1350, Section 5.
// Opcodes are used to send different types of messages between a client and
// a server.
type Opcode uint16

// Opcode constants taken from RFC 1350, Section 5.
const (
	OpcodeRead  Opcode = 1
	OpcodeWrite Opcode = 2
	OpcodeError Opcode = 5

	// Opcode types only used for internal communication
	opcodeDATA Opcode = 3
	opcodeACK  Opcode = 4
)

// Mode represents a TFTP transfer mode, as defined in RFC 1350, Section 1.
// Modes are used to negotiate different types of transfer methods between a
// client and a server.
type Mode string

// Mode constants taken from RFC 1350, Section 1.
//
// The mail mode is intentionally omitted, per RFC 1350:
//   The mail mode is obsolete and should not be implemented or used.
const (
	ModeNetASCII Mode = "netascii"
	ModeOctet    Mode = "octet"
)

// ErrorCode represents a TFTP error code, as defined in RFC 1350, Appendix I.
// ErrorCodes are used to communicate different types of errors between a
// client and a server.
type ErrorCode uint16

// ErrorCode constants taken from RFC 1350, Appendix I.
const (
	ErrorCodeUndefined         ErrorCode = 0
	ErrorCodeFileNotFound      ErrorCode = 1
	ErrorCodeAccessViolation   ErrorCode = 2
	ErrorCodeDiskFull          ErrorCode = 3
	ErrorCodeIllegalOperation  ErrorCode = 4
	ErrorCodeUnknownTransferID ErrorCode = 5
	ErrorCodeFileExists        ErrorCode = 6
	ErrorCodeNoSuchUser        ErrorCode = 7
)

// Handler provides an interface which allows structs to act as TFTP server
// handlers.  ServeTFTP implementations receive a copy of the incoming TFTP
// request via the Request parameter, and allow outgoing communication via
// the ResponseWriter.
type Handler interface {
	ServeTFTP(ResponseWriter, *Request)
}

// HandlerFunc is an adapter type which allows the use of normal functions as
// TFTP handlers.  If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler struct that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeTFTP calls f(w, r), allowing regular functions to implement Handler.
func (f HandlerFunc) ServeTFTP(w ResponseWriter, r *Request) {
	f(w, r)
}

// ResponseWriter provides an interface which allows a TFTP handler to write
// TFTP data packets to a client.  The default ResponseWriter binds a new UDP
// socket to communicate with a client, and closes it when Close is called.
//
// ResponseWriter implementations should buffer some data internally, in order
// to send 512 byte blocks in a "lock-step" fashion to a client.
type ResponseWriter interface {
	// Write implements io.Writer, and allows raw data to be sent to a client.
	Write([]byte) (int, error)

	// Close closes the underlying UDP socket used to communicate with a
	// client.
	Close() error

	// Flush flushes any buffered data to a client, signaling the end of data
	// transfer.
	Flush() error
}

// fromNetASCII performs the necessary conversions from an input buffer
// needed when a client is using netascii mode.
//
// BUG(mdlayher): this could be heavily optimized in the future to not read
// a single byte at a time
func fromNetASCII(p []byte) []byte {
	b := make([]byte, len(p))
	copy(b, p)

	// If using netascii mode, some conversions must be made from
	// input data:
	//   -   CR+LF -> LF
	//   - CR+NULL -> CR
	b = bytes.Replace(b, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	b = bytes.Replace(b, []byte{'\r', 0}, []byte{'\r'}, -1)

	return b
}
