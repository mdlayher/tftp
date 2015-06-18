package tftp

import (
	"net"
)

// Request represents a processed TFTP request received by a server.
// Its struct members contain information regarding the request type, filename,
// transfer mode, etc.
type Request struct {
	// Opcode specifies the requested operation, such as read or write.
	Opcode Opcode

	// Filename specifies the requested filename for a read request, or where a
	// file should be written on the server for a write request.
	Filename string

	// Mode specifies the transfer mode for this request, such as netascii or
	// octet.
	Mode Mode

	// Length of the TFTP request, in bytes.
	Length int64

	// Network address which was used to contact the TFTP server.  The server
	// will automatically set up a socket to communicate with this address.
	RemoteAddr string
}

// parseRequest creates a new Request from an input byte slice and UDP address.
// It populates the basic struct members which can be used in a TFTP handler.
//
// If the input byte slice is not a valid TFTP request packet, errInvalidRequestPacket
// is returned.
func parseRequest(b []byte, remoteAddr net.Addr) (*Request, error) {
	p, err := parseRequestPacket(b)
	if err != nil {
		return nil, err
	}

	return &Request{
		Opcode:     p.Opcode,
		Filename:   p.Filename,
		Mode:       p.Mode,
		Length:     int64(len(b)),
		RemoteAddr: remoteAddr.String(),
	}, nil
}
