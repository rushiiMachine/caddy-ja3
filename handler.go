package caddy_ja3

import (
	"errors"
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
	if !ctx.AppIsConfigured(CacheAppId) {
		return errors.New("global cache is not configured")
	}
	a, err := ctx.App(CacheAppId)
	if err != nil {
		return err
	}

	h.cache = a.(*Cache)
	h.log = ctx.Logger(h)
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (h *JA3Handler) UnmarshalCaddyfile(_ *caddyfile.Dispenser) error {
	// no-op impl
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (h *JA3Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	if req.TLS.HandshakeComplete {
		ja3 := h.cache.GetJA3(req.RemoteAddr)

		if ja3 == nil {
			h.log.Error("ClientHello missing from cache for " + req.RemoteAddr)
		} else {
			h.log.Debug("Attaching JA3 to request for " + req.RemoteAddr)
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
