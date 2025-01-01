package caddy_ja3

import (
	"encoding/binary"
	"errors"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"io"
	"net"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JA3ListenerWrapper{})
}

type JA3ListenerWrapper struct {
	cache *Cache
	log   *zap.Logger
}

type clientHelloListener struct {
	net.Listener
	cache *Cache
	log   *zap.Logger
}

type clientHelloConnListener struct {
	net.Conn
	cache *Cache
	log   *zap.Logger
}

// CaddyModule implements caddy.Module
func (JA3ListenerWrapper) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.listeners.ja3",
		New: func() caddy.Module { return new(JA3ListenerWrapper) },
	}
}

func (l *JA3ListenerWrapper) Provision(ctx caddy.Context) error {
	a, err := ctx.App(CacheAppId)
	if err != nil {
		return err
	}

	l.cache = a.(*Cache)
	l.log = ctx.Logger(l)

	// Disable TLS session resumption via session tickets
	app, err := ctx.App("tls")
	if err != nil {
		return err
	}
	tlsApp := app.(*caddytls.TLS)
	if tlsApp.SessionTickets == nil {
		tlsApp.SessionTickets = new(caddytls.SessionTicketService)
	}
	tlsApp.SessionTickets.Disabled = true
	l.log.Debug("adjusted config: disabled TLS session tickets")

	return nil
}

// WrapListener implements caddy.ListenerWrapper
func (l *JA3ListenerWrapper) WrapListener(ln net.Listener) net.Listener {
	return &clientHelloListener{
		ln,
		l.cache,
		l.log,
	}
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (l *JA3ListenerWrapper) UnmarshalCaddyfile(_ *caddyfile.Dispenser) error {
	// no-op impl
	return nil
}

// Accept implements net.Listener
func (l *clientHelloListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return conn, err
	}

	raw, err := ReadClientHello(conn)
	if err == nil {
		addr := conn.RemoteAddr().String()
		if err := l.cache.SetClientHello(addr, raw); err != nil {
			l.log.Error("Failed to cache JA3 for connection", zap.String("addr", addr), zap.Error(err))
		}

		l.log.Debug("Cached JA3 for connection", zap.String("addr", conn.RemoteAddr().String()))
	} else {
		l.log.Debug("Failed to read ClientHello from connection", zap.String("addr", conn.RemoteAddr().String()), zap.Error(err))
	}

	return RewindConn(&clientHelloConnListener{
		conn,
		l.cache,
		l.log,
	}, raw)
}

// Close implements net.Conn
func (l *clientHelloConnListener) Close() error {
	addr := l.Conn.RemoteAddr().String()

	l.cache.ClearJA3(addr)
	l.log.Debug("Disposing of JA3 for connection", zap.String("addr", addr))

	return l.Conn.Close()
}

// ReadClientHello reads as much of a ClientHello as possible and returns it.
// If any error was encountered, then an error is returned as well and the raw bytes are not a full ClientHello.
func ReadClientHello(r io.Reader) (raw []byte, err error) {
	// Obtained from https://github.com/gaukas/clienthellod/blob/7cce34b88b314256c8759998f6192860f6f6ede5/clienthello.go#L68

	// Read a TLS record
	// Read exactly 5 bytes from the reader
	raw = make([]byte, 5)
	if _, err = io.ReadFull(r, raw); err != nil {
		return
	}

	// Check if the first byte is 0x16 (TLS Handshake)
	if raw[0] != 0x16 {
		err = errors.New("not a TLS handshake record")
		return
	}

	// Read exactly length bytes from the reader
	raw = append(raw, make([]byte, binary.BigEndian.Uint16(raw[3:5]))...)
	_, err = io.ReadFull(r, raw[5:])
	return raw, err
}

// Interface guards
var (
	_ caddy.Provisioner     = (*JA3ListenerWrapper)(nil)
	_ caddy.ListenerWrapper = (*JA3ListenerWrapper)(nil)
	_ caddyfile.Unmarshaler = (*JA3ListenerWrapper)(nil)
)
