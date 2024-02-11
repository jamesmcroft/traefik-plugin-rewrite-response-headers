# Rewrite Response Headers

Rewrite response headers is a middleware plugin for Traefik which rewrites the HTTP response headers by replacing a search regex by a replacement string.

## Configuration

### Static

```yaml
pilot:
  token: "xxxx"

experimental:
  plugins:
    rewriteResponseHeaders:
      moduleName: github.com/jamesmcroft/traefik-plugin-rewrite-response-headers
      version: "v1.0.0"
```

### Dynamic

To configure the Rewrite Response Headers plugin you should create a [middleware](https://docs.traefik.io/middlewares/overview/) in your dynamic configuration as explained [here](https://doc.traefik.io/traefik/middlewares/overview/). The following example creates and uses the `rewriteResponseHeaders` middleware plugin to replace a custom `Operation-Location` header which provides a `http://` URL with a `https://` URL.

```yaml
http:
  services:
    serviceRoot:
      loadBalancer:
        servers:
          - url: "http://localhost:8080"

  middlewares:
    rewrite-operation-location-header:
      plugin:
        rewriteResponseHeaders:
          rewrites:
            - header: "Operation-Location"
              regex: "^http://(.+?)/(.+)$"
              replacement: "https://{{RequestHost}}/$2"

  routers:
    routerRoot:
      rule: "PathPrefix(`/`)"
      service: "serviceRoot"
      middlewares:
        - "rewrite-operation-location-header"
```

> [!NOTE]
> This plugin includes a `{{RequestHost}}` token which can be used in the `replacement` string to include the original request host in the replaced header value. **It is not required to use this token**.
