package tftp

import (
	"net"
)

// Server represents a TFTP server, and is used to configure a TFTP server's
// behavior.
type Server struct {
	// Addr is the network address which this server should bind to.  The
	// default value is :69, as specified in RFC 1350, Section 4.
	Addr string

	// Handler is the handler to use while serving TFP requests.
	// Handler must not be nil.
	Handler Handler
}

// ListenAndServe listens for UDP connections on the specified address, using
// the default Server configuration and specified handler to handle TFTP
// connections.
func ListenAndServe(addr string, handler Handler) error {
	return (&Server{
		Addr:    addr,
		Handler: handler,
	}).ListenAndServe()
}

// ListenAndServe listens on the address specified by s.Addr.  Serve is called
// to handle serving TFTP traffic once ListenAndServe opens a UDP packet
// connection.
func (s *Server) ListenAndServe() error {
	conn, err := net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}

	defer conn.Close()
	return s.Serve(conn)
}

// Serve configures and accepts incoming connections on PacketConn p, creating a
// new goroutine for each.
//
// The service goroutine reads requests, generate the appropriate Request and
// ResponseWriter values, then calls s.Handler to handle the request.
func (s *Server) Serve(p net.PacketConn) error {
	// RRQ and WRQ packets are received here before creating a goroutine to
	// handle data transfer.  There appears to be no maximum limit for the
	// size of one of these packets, so we will go with the Ethernet MTU,
	// since TFTP packets must fit inside one, unfragmented IP packet.
	buf := make([]byte, 1500)
	for {
		n, addr, err := p.ReadFrom(buf)
		if err != nil {
			return err
		}

		go s.newConn(addr, n, buf).serve()
	}
}

// conn represents an in-flight TFTP connection, and contains information about
// the connection and server.
type conn struct {
	conn       net.PacketConn
	remoteAddr net.Addr
	server     *Server
	buf        []byte
}

// newConn creates a new conn using information received in a single TFTP
// request.  newConn makes a copy of the input buffer for use in handling
// a single request.
//
// BUG(mdlayher): consider using a sync.Pool with many buffers available to avoid
// allocating a new one on each request.
func (s *Server) newConn(addr net.Addr, n int, buf []byte) *conn {
	c := &conn{
		remoteAddr: addr,
		server:     s,
		buf:        make([]byte, n),
	}
	copy(c.buf, buf[:n])

	return c
}

// serve handles serving an individual TFTP request, and is invoked in a
// goroutine.
func (c *conn) serve() {
	// Attempt to parse a Request from a raw packet, providing a nicer
	// API for callers to implement their own TFTP request handlers
	r, err := parseRequest(c.buf, c.remoteAddr)
	if err != nil {
		// BUG(mdlayher): send ERROR response on invalid request
		if err == errInvalidRequestPacket {
			return
		}

		return
	}

	// Set up response by binding a new UDP socket to handle this request
	w, err := newResponse(c.server.Addr, c.remoteAddr, r.Mode)
	if err != nil {
		return
	}

	// This will panic if Handler is nil.
	// TODO(mdlayher): determine if a ServeMux type would be useful.
	c.server.Handler.ServeTFTP(w, r)
}
