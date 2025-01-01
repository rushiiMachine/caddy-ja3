package caddy_ja3

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

const (
	ConfigAppId = "ja3.config"
)

type Config struct {
	// This safeguards against TLS order randomization in Chrome by
	// ensuring TLS extensions are always in the same order to ensure
	// a consistent fingerprint in the end.
	SortExtensions bool `json:"sort_extensions"`
}

func init() {
	caddy.RegisterModule(Config{})
	httpcaddyfile.RegisterGlobalOption("ja3", parseCaddyfile)
}

func parseCaddyfile(d *caddyfile.Dispenser, _ any) (any, error) {
	var config Config

	for d.Next() {
		for d.NextBlock(0) {
			opt := d.Val()

			switch opt {
			case "sort_extensions":
				config.SortExtensions = true

			default:
				return nil, d.Errf("unrecognized directive: %s", opt)
			}
		}
	}

	return httpcaddyfile.App{
		Name:  ConfigAppId,
		Value: caddyconfig.JSON(config, nil),
	}, nil
}

// CaddyModule implements caddy.Module
func (Config) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  ConfigAppId,
		New: func() caddy.Module { return new(Config) },
	}
}

// Start implements caddy.App
func (c *Config) Start() error {
	return nil
}

// Stop implements caddy.App
func (c *Config) Stop() error {
	return nil
}

// Interface guards
var (
	_ caddy.App    = (*Config)(nil)
	_ caddy.Module = (*Config)(nil)
)
