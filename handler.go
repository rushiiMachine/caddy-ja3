package caddy_ja3

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JA3Handler{})
	httpcaddyfile.RegisterHandlerDirective("ja3", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		handler := &JA3Handler{}
		return handler, handler.UnmarshalCaddyfile(h.Dispenser)
	})
}

type JA3Handler struct {
	cache *Cache
	log   *zap.Logger
}

func (JA3Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.ja3",
		New: func() caddy.Module { return new(JA3Handler) },
	}
}

// Provision implements caddy.Provisioner
func (h *JA3Handler) Provision(ctx caddy.Context) error {
	a, err := ctx.App(CacheAppId)
	if err != nil {
		return err
	}

	h.cache = a.(*Cache)
	h.log = ctx.Logger(h)
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (h *JA3Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// no-op impl
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (h *JA3Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	if req.TLS.HandshakeComplete && req.ProtoMajor < 3 { // Check that this uses TLS and < HTTP/3
		ja3 := h.cache.GetJA3(req.RemoteAddr)

		if ja3 == nil {
			h.log.Error("ClientHello missing from cache", zap.String("addr", req.RemoteAddr))
		} else {
			h.log.Debug("Attaching JA3 to request", zap.String("addr", req.RemoteAddr))
			req.Header.Add("JA3", *ja3)
		}
	}

	return next.ServeHTTP(rw, req)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*JA3Handler)(nil)
	_ caddyhttp.MiddlewareHandler = (*JA3Handler)(nil)
	_ caddyfile.Unmarshaler       = (*JA3Handler)(nil)
)
