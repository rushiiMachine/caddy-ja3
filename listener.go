package caddy_ja3

import (
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JA3Listener{})
}

type JA3Listener struct {
	cache *Cache
	log   *zap.Logger
}

type tlsClientHelloListener struct {
	net.Listener
	cache *Cache
	log   *zap.Logger
}

// CaddyModule implements caddy.Module
func (JA3Listener) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.listeners.ja3",
		New: func() caddy.Module { return new(JA3Listener) },
	}
}

func (l *JA3Listener) Provision(ctx caddy.Context) error {
	if !ctx.AppIsConfigured(CacheAppId) {
		return errors.New("global cache is not configured")
	}
	a, err := ctx.App(CacheAppId)
	if err != nil {
		return err
	}

	l.cache = a.(*Cache)
	l.log = ctx.Logger(l)
	return nil
}

// WrapListener implements caddy.ListenerWrapper
func (l *JA3Listener) WrapListener(ln net.Listener) net.Listener {
	return &tlsClientHelloListener{
		ln,
		l.cache,
		l.log,
	}
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (l *JA3Listener) UnmarshalCaddyfile(_ *caddyfile.Dispenser) error {
	// no-op impl
	return nil
}

// Accept implements net.Listener
func (l *tlsClientHelloListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return conn, err
	}

	ch, err := ReadClientHello(conn)
	if err == nil {
		addr := conn.RemoteAddr().String()
		if err := l.cache.SetClientHello(addr, ch); err != nil {
			l.log.Error("Failed to cache JA3 for "+addr, zap.Error(err))
		}

		l.log.Debug("Cached JA3 for " + conn.RemoteAddr().String())
	} else {
		l.log.Debug("Failed to read ClientHello for "+conn.RemoteAddr().String(), zap.Error(err))
	}

	return RewindConn(conn, ch)
}

// TODO: fix cache leak
// Close implements net.Listener
//func (l *tlsClientHelloListener) Close() error {
//	addr := l.Listener.Addr().String()
//
//	l.cache.ClearJA3(addr)
//	l.log.Debug("Disposing of JA3 for" + addr)
//
//	return nil
//}

func ReadClientHello(r io.Reader) (ch []byte, err error) {
	// Obtained from https://github.com/gaukas/clienthellod/blob/7cce34b88b314256c8759998f6192860f6f6ede5/clienthello.go#L68

	// Read a TLS record
	// Read exactly 5 bytes from the reader
	raw := make([]byte, 5)
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
	return raw, nil
}

// Interface guards
var (
	_ caddy.Provisioner     = (*JA3Listener)(nil)
	_ caddy.ListenerWrapper = (*JA3Listener)(nil)
	_ caddyfile.Unmarshaler = (*JA3Listener)(nil)
)
