// Command rotftpd is a simple, read-only, TFTP server, which can be used to
// serve files from a single directory over TFTP.
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mdlayher/tftp"
)

var (
	// addr specifies the address the TFTP server will listen on.
	addr = flag.String("addr", ":69", "host:port pair for TFTP server to listen on")

	// dir specifies the directory the TFTP server will serve.
	dir = flag.String("dir", ".", "directory containing files to be served over TFTP")
)

func main() {
	flag.Parse()

	d := filepath.Clean(*dir)
	h := &Handler{
		Directory: d,
	}

	log.Printf("serving TFTP directory %q on %s", d, *addr)
	if err := tftp.ListenAndServe(*addr, h); err != nil {
		log.Fatal(err)
	}
}

// Handler is a simple tftp.Handler implementation.
type Handler struct {
	Directory string
}

// ServeTFTP serves incoming TFTP using the directory specified in Handler.
func (h *Handler) ServeTFTP(w tftp.ResponseWriter, r *tftp.Request) {
	// Strip any directories from filename
	r.Filename = filepath.Base(r.Filename)

	// Ignore write requests
	if r.Opcode == tftp.OpcodeWrite {
		log.Printf("ignoring: [%s] %q (server is read-only)", r.RemoteAddr, r.Filename)

		// BUG(mdlayher): should report an error of some kind to the client so
		// it does not hang forever.
		return
	}

	// Open file to begin write
	f, err := os.Open(filepath.Join(h.Directory, r.Filename))
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	// Check file's size
	s, err := f.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf(" serving: [%s] %q, %d bytes", r.RemoteAddr, r.Filename, s.Size())
	start := time.Now()

	// Begin copying file to client
	if _, err := io.Copy(w, f); err != nil && err != io.EOF {
		log.Println(err)
		return
	}

	// Flush any remaining buffered bytes
	if err := w.Flush(); err != nil {
		log.Println(err)
		return
	}

	// Close server's socket for this client
	_ = w.Close()
	log.Printf("complete: [%s] %q, %d bytes in %s", r.RemoteAddr, r.Filename, s.Size(), time.Now().Sub(start))
}
