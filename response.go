package tftp

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"net"
	"time"
)

const (
	// blockSize is the RFC 1350 specified DATA packet size for a
	// single read or write.
	blockSize = 512
)

// response is the default ResponseWriter implementation.  It performs some
// internal buffering, and if needed, netascii conversions, to write DATA
// packets to a client.
type response struct {
	ResponseWriter
}

// newResponse creates a new response, setting up a UDP socket to perform
// communication for a single client.
func newResponse(serverAddr string, remoteAddr net.Addr, mode Mode) (*response, error) {
	host, _, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, err
	}

	// Bind to a system-assigned UDP port using the server's address
	conn, err := net.ListenPacket("udp", net.JoinHostPort(host, "0"))
	if err != nil {
		return nil, err
	}

	// Set up writer which communicates via socket and buffers input
	// appropriately for TFTP
	bsw := &bufferedSocketResponseWriter{
		conn:       conn,
		remoteAddr: remoteAddr,

		buf: bytes.NewBuffer(nil),

		rb: make([]byte, blockSize+4),
		wb: make([]byte, blockSize+4),
	}

	// If using netascii mode, wrap with ResponseWriter which seamlessly
	// converts writes to netascii
	var rw ResponseWriter = bsw
	if mode == ModeNetASCII {
		rw = &netASCIIResponseWriter{bsw}
	}

	return &response{
		ResponseWriter: rw,
	}, nil
}

// bufferedSocketResponseWriter is a ResponseWriter which communicates with a
// TFTP client over a socket, and buffers data internally.
type bufferedSocketResponseWriter struct {
	// Connection and address used to communicate with a client
	conn       net.PacketConn
	remoteAddr net.Addr

	// Reusable read and write buffers
	rb []byte
	wb []byte

	// Buffer to store blocks which are not large enough to be written
	buf *bytes.Buffer

	// Current block number
	block uint16
}

// Write implements io.Writer, and performs internal buffering of data to
// communicate with a client.  Write attempts to send as many available blocks
// as possible when called, buffering any excess data for future writes.
func (w *bufferedSocketResponseWriter) Write(p []byte) (int, error) {
	// Store data in buffer to be output in blocks
	// (never returns an error, per documentation)
	n, _ := w.buf.Write(p)

	// If buffer and input bytes cannot create an entire block, wait until
	// next call or flush before performing any writes
	if w.buf.Len() < blockSize {
		return n, nil
	}

	// Calculate how many blocks we can send with the data currently
	// in the buffer
	available := int(math.Floor(
		float64(w.buf.Len()) / float64(blockSize),
	))

	// Flush as many available blocks as possible
	for i := 0; i < available; i++ {
		if err := w.writeOneBlock(); err != nil {
			return n, err
		}
	}

	return n, nil
}

// Close closes the underlying socket used to communicate with a client.
func (w *bufferedSocketResponseWriter) Close() error {
	return w.conn.Close()
}

// Flush writes up to a single block of data to a client.  Flush should only
// be called once an io.Reader returns EOF, in order to ensure that every byte
// from the Reader is flushed to the client.
func (w *bufferedSocketResponseWriter) Flush() error {
	return w.writeOneBlock()
}

// writeOneBlock attempts to write a single block of data to a client, and
// waits for acknowledgement or an error in reply.
//
// BUG(mdlayher): break this up into smaller methods
func (w *bufferedSocketResponseWriter) writeOneBlock() error {
	// Write data header with incremented block number and send
	// one block to client
	w.block++
	binary.BigEndian.PutUint16(w.wb[0:2], uint16(opcodeDATA))
	binary.BigEndian.PutUint16(w.wb[2:4], w.block)

	// Copy up to blockSize bytes into write buffer for a single write
	// transaction
	cn := copy(w.wb[4:], w.buf.Next(blockSize))

	for {
		// Set timeouts for a reasonable amount of time before retrying
		if err := w.conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			return err
		}

		// Write block to client using its connection, ensure that the
		// correct number of bytes were written
		wn, err := w.conn.WriteTo(w.wb[:cn+4], w.remoteAddr)
		if err != nil {
			// Allow retries on timeout
			oerr, ok := err.(*net.OpError)
			if !ok {
				return err
			}

			if oerr.Timeout() {
				continue
			}

			return err
		}
		if wn != cn+4 {
			return io.ErrShortWrite
		}

		// Wait for ACK or ERROR response from client
		var readN int
		for {
			rn, addr, err := w.conn.ReadFrom(w.rb)
			if err != nil {
				// Allow retries on timeout
				oerr, ok := err.(*net.OpError)
				if !ok {
					return err
				}

				if oerr.Timeout() {
					continue
				}

				return err
			}

			// BUG(mdlayher): send errors for wrong TID if an unknown
			// client starts communicating on this port
			_ = addr

			readN = rn
			break
		}

		// Parse ACK or ERROR packet
		ack, err := parseACKPacket(w.rb[:readN])
		if err != nil {
			return err
		}

		// If client reports the previous block as acknowledged again, we
		// must repeat the process
		if ack.Block == w.block-1 {
			continue
		}

		return nil
	}
}
