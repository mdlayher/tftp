tftp [![Build Status](https://travis-ci.org/mdlayher/tftp.svg?branch=master)](https://travis-ci.org/mdlayher/tftp) [![Coverage Status](https://coveralls.io/repos/mdlayher/tftp/badge.svg?branch=master)](https://coveralls.io/r/mdlayher/tftp?branch=master) [![GoDoc](http://godoc.org/github.com/mdlayher/tftp?status.svg)](http://godoc.org/github.com/mdlayher/tftp)
====

Package `tftp` implements a TFTP server, as described in IETF RFC 1350.  MIT Licensed.

A basic, read-only, TFTP server (`rotftpd`) which demonstrates the library is
available at [cmd/rotftpd](https://github.com/mdlayher/tftp/blob/master/cmd/rotftpd).

At this time, the API is not stable, and may change over time.  The eventual
goal is to implement a server, client, and testing facilities for consumers
of this package.

The design of this package is inspired by Go's `net/http` package.  The Go
standard library is Copyright (c) 2012 The Go Authors. All rights reserved.
The Go license can be found at https://golang.org/LICENSE.
