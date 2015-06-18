rotftpd
=======

Command `rotftpd` is a simple, read-only, TFTP server, which can be used to
serve files from a single directory over TFTP.

Usage
=====

```
$ ./rotftpd -h
Usage of ./rotftpd:
  -addr=":69": host:port pair for TFTP server to listen on
  -dir=".": directory containing files to be served over TFTP
```
