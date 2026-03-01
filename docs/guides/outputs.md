# Defining Outputs

Zorya serializes your output struct into an HTTP response automatically. The struct shape controls the status code, response headers, and body.

## Output struct shape

```go
type MyOutput struct {
    // Status overrides the HTTP status code. Tag with json:"-" so it is
    // not serialized into the response body.
    Status int `json:"-"`

    // Fields at the top level (non-Body, non-Status) are written as
    // response headers. The struct field name becomes the header name.
    Location   string `header:"Location"`
    ETag       string `header:"ETag"`

    // Body is the response body. It is serialized according to the
    // negotiated content type.
    Body struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
    }
}
```

## Status codes

Set `Status int \`json:"-"\`` to return a custom status code.

```go
return &CreateUserOutput{Status: http.StatusCreated, ...}, nil
```

If `Status` is omitted or zero, Zorya uses `200 OK` for handlers that return a body and `204 No Content` for handlers whose `Body` is empty.

### Dynamic status with StatusProvider

Implement `StatusProvider` to compute the status at runtime:

```go
type StatusProvider interface {
    GetStatus() int
}
```

`ErrorModel` already implements this interface, so returning an error automatically carries the right status code.

## Response headers

Declare top-level struct fields (outside `Body`) to write response headers. The field name is used as-is unless a `header` tag is present.

```go
type CreateUserOutput struct {
    Status   int    `json:"-"`
    Location string `header:"Location"`
    Body     struct {
        ID int `json:"id"`
    }
}

// In the handler:
return &CreateUserOutput{
    Status:   http.StatusCreated,
    Location: fmt.Sprintf("/users/%d", user.ID),
    Body:     struct{ ID int `json:"id"` }{ID: user.ID},
}, nil
```

## Empty body (204)

For responses with no body (DELETE, some PUT), omit the `Body` field entirely:

```go
type DeleteUserOutput struct{}

func deleteUser(ctx context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
    // ... delete logic ...
    return &DeleteUserOutput{}, nil // Zorya sends 204 No Content
}
```

## Streaming body

Set `Body` to a `func(w http.ResponseWriter) error` to take full control of writing the response. Zorya calls the function and passes through any error.

```go
type StreamOutput struct {
    Body func(w http.ResponseWriter) error
}

out := &StreamOutput{}
out.Body = func(w http.ResponseWriter) error {
    w.Header().Set("Content-Type", "text/event-stream")
    for i := range 5 {
        fmt.Fprintf(w, "data: event %d\n\n", i)
        w.(http.Flusher).Flush()
        time.Sleep(time.Second)
    }
    return nil
}
return out, nil
```

See [Streaming (SSE)](streaming.md) for the full guide.

## Raw bytes body

Set `Body` to `[]byte` to write raw bytes without serialization:

```go
type PDFOutput struct {
    Status int    `json:"-"`
    Body   []byte
}
```

Zorya writes the bytes directly with the negotiated `Content-Type`.
