<p align="center">
  <a href="https://github.com/fwiedmann/differ">
    <img src="image/prox.png" width=100 height=100>
  </a>

  <h3 align="center">prox</h3>

  <p align="center">
     A simple reverse proxy
  </p>
</p>

 ![love](https://img.shields.io/badge/made%20with-%E2%9D%A4%EF%B8%8F-lightgrey)

## Features:

- HTTP Proxy
- In-Memory Cache
- Dynamic Route reload
- Dynamic Route reload
- Health Endpoint
- Metrics
- Middlewares:
    - HTTPs redirect
    - Forward Host Address

### Test & Build

```bash
    go test -coverprofile testcov -race ./...
    CGO_ENABLED=0 go build -o prox cmd/main.go
```


## Run
```bash
Usage:
  prox [flags]

Flags:
  -h, --help                   help for prox
      --loglevel string        Set a log level (default "info")
      --routes-config string   Path to routes config file (default "routes.yaml")
      --static-config string   Path to static config file (default "static.yaml")
      --tls-config string      Path to routes tls file (default "tls.yaml")
```




## Configuration

### Static Configuration

The static configuration will initial configure `prox`.

```yaml
access-log-enabled: true # optional,default false
infra-port: 9100 # optional,default 9100
cache:
  enabled: true # optional,default false
  cache-max-size-in-mega-byte: 10000  # optional,default -1 which means infinite
ports:
  - name: "http" # required
    port: 80 # required
    tls: false # optional,default false
  - name: "https"
    port: 443
    tls: true # optional,default false
```

### Dynamic Route Configuration

The dynamic route config includes all `prox` routes which will be used for incoming http traffic. On config changes `prox` will reload and validate the new configuration.
Note that route `name` has to be an unique identifier.

```yaml
- name: "backend-1-http" # required
  cache-enabled: true # optional, default false
  cache-timeout: "5m" # optional, default 10m
  cache-max-body-size-in-mb: 100 # optional, default -1 which means infinite
  upstream-url: "https://docker.com" # required
  upstream-timeout: "20s" # optional, default 10s
  upstream-skip-tls: false # optional, default false
  priority: 3 # optional, default false
  port: 80 # required
  hostname: "example.com" # required
  middlewares:
    https-redirect-enabled: true # optional, default false
    https-redirect-port: 443 # optional, default 433 only when "https-redirect-enabled: true"
    forward-host-header: true  # optional, default false

- name: "backend-1-https"
  cache-enabled: true
  cache-timeout: "5m"
  cache-max-body-size-in-mb: 100
  upstream-url: "https://docker.com"
  upstream-timeout: "20s"
  upstream-skip-tls: false
  priority: 3
  port: 443
  hostname: "example.com"
  middlewares:
    forward-host-header: true
    https-redirect-enabled: true
    https-redirect-port: 443
```

### Dynamic TLS Configuration

The dynamic TLS configuration dynamically load the available TLS certificates for the `prox` ports, with the `tls: true` option set, from the given file paths in the config file.

```yaml
- certificate: "/certs/localhost.pem" # required
  key: "/certs/localhost-key.pem"   # required
- certificate: "/certs/test.pem" # required
  key: "/certs/test-key.pem"   # required
```
