# Content Negotiation

Zorya automatically negotiates the response format based on the client's `Accept` header. JSON and CBOR are supported by default.

## Default formats

| Content-Type | Enabled | Notes |
|---|---|---|
| `application/json` | Yes | Default when `Accept` is absent or `*/*` |
| `application/cbor` | Yes | [RFC 7049](https://www.rfc-editor.org/rfc/rfc7049) binary serialization |

## How it works

1. Zorya reads the `Accept` header from the request.
2. It selects the best match from the registered formats.
3. It serializes the response body using the selected format.
4. It sets the `Content-Type` response header accordingly.

When no match is found and `NoFormatFallback` is false (default), Zorya falls back to `application/json`.

## Requesting CBOR

```bash
curl -H "Accept: application/cbor" http://localhost:8080/users/1
```

## Adding a custom format

Implement a `Format` and register it:

```go
import "github.com/talav/zorya"

xmlFormat := zorya.Format{
    Marshal: func(w io.Writer, v any) error {
        return xml.NewEncoder(w).Encode(v)
    },
}

api := zorya.NewAPI(adapter, zorya.WithFormat("application/xml", xmlFormat))
```

### Replacing all formats

To serve only your custom formats (disabling JSON and CBOR):

```go
api := zorya.NewAPI(adapter, zorya.WithFormatsReplace(map[string]zorya.Format{
    "application/xml": xmlFormat,
}))
```

## ContentTypeProvider

Implement `ContentTypeProvider` on a response body type to override its content type regardless of negotiation. This is used internally by `ErrorModel` to always return `application/problem+json`.

```go
type ContentTypeProvider interface {
    ContentType(negotiated string) string
}
```

Example:

```go
type PDFResponse struct {
    Data []byte
}

func (p *PDFResponse) ContentType(_ string) string {
    return "application/pdf"
}
```

## Strict mode

Set `NoFormatFallback: true` to return `406 Not Acceptable` when no format matches the `Accept` header instead of falling back to JSON:

```go
api := zorya.NewAPI(adapter, zorya.WithConfig(&zorya.Config{
    NoFormatFallback: true,
}))
```

## Default format

Override the fallback format (used when `Accept` is absent):

```go
api := zorya.NewAPI(adapter, zorya.WithDefaultFormat("application/cbor"))
```
