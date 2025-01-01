# caddy-ja3

A caddy plugin to get JA3 fingerprints from requests as a header.

## Building with xcaddy

```shell
xcaddy build \
  --with github.com/rushiiMachine/caddy-ja3
```

## Sample Caddyfile

Note that this enforces HTTPS (TLS).\
You can add a http_redirect to automatically redirect `http` -> `https` like shown below.

TLS `ClientHello`s do not exist on HTTP/3 connections.
No `ja3` header will be present on such requests.
Unless another way is used to fingerprint HTTP/3 aka. QUIC connections, it's recommended to disable HTTP/3.

This module also disables TLS session resumption globally to always retrieve a full `ClientHello`.
This is done through the usage of
[caddytls's `session_tickets/disabled`](https://caddyserver.com/docs/modules/tls#session_tickets/disabled)
config option internally.

```caddyfile
{
    # If using a different responder like reverse_proxy, change this accordingly
    order ja3 before respond
    servers {
        # Disable HTTP/3
        protocols h1 h2

        listener_wrappers {
            http_redirect
            ja3
            tls
        }
    }
}

localhost {
    ja3
    # Configure your TLS however you want
    tls internal
    # JA3 fingerprint is added to the request as the "JA3" header
    respond "Your JA3: {header.ja3}"
}
```
