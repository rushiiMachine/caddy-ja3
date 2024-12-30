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

```
{
  order ja3 before respond # change this to whatever idk
  servers {
     listener_wrappers {
       http_redirect
       ja3
       tls
     }
  }
}

localhost:2020 {
  ja3
  tls internal                     # Configure your tls however you want
  respond "Your JA3: {header.ja3}" # JA3 is added to the request as a header ("ja3")
}
```

## Disclaimer

I am not guaranteeing you 100% uptime\
This should work but feel free to file an issue
