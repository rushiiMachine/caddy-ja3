package caddy_ja3

// Obtained from https://github.com/gaukas/clienthellod/blob/7cce34b88b314256c8759998f6192860f6f6ede5/internal/utils/rewindconn.go

import (
	"bytes"
	"errors"
	"io"
	"net"
)

type rewindConn struct {
	net.Conn
	reader bytes.Reader
}

func RewindConn(c net.Conn, buf []byte) (net.Conn, error) {
	if c == nil {
		return nil, errors.New("cannot rewind nil connection")
	}

	if len(buf) == 0 {
		return c, nil
	}

	return &rewindConn{
		Conn:   c,
		reader: *bytes.NewReader(buf),
	}, nil
}

// Read is ...
func (c *rewindConn) Read(b []byte) (int, error) {
	if c.reader.Size() == 0 {
		return c.Conn.Read(b)
	}
	n, err := c.reader.Read(b)
	if errors.Is(err, io.EOF) {
		c.reader.Reset([]byte{})
		return n, nil
	}
	return n, err
}

// CloseWrite is ...
func (c *rewindConn) CloseWrite() error {
	if cc, ok := c.Conn.(*net.TCPConn); ok {
		return cc.CloseWrite()
	}
	if cw, ok := c.Conn.(interface {
		CloseWrite() error
	}); ok {
		return cw.CloseWrite()
	}
	return errors.New("not supported")
}

// Interface guards
var (
	_ net.Conn = (*rewindConn)(nil)
)
