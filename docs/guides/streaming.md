# Streaming (SSE)

Zorya supports streaming responses by setting the `Body` field to a function. This gives the handler full control over writing to the response, enabling Server-Sent Events (SSE), chunked JSON, and any other streaming pattern.

## Streaming body signature

```go
type StreamingOutput struct {
    Body func(w http.ResponseWriter) error
}
```

When Zorya encounters a `func(http.ResponseWriter) error` body, it calls the function directly. The function is responsible for setting headers and writing all data.

## Server-Sent Events (SSE)

```go
type EventStreamOutput struct {
    Body func(w http.ResponseWriter) error
}

func streamEvents(ctx context.Context, input *struct{}) (*EventStreamOutput, error) {
    out := &EventStreamOutput{}
    out.Body = func(w http.ResponseWriter) error {
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.WriteHeader(http.StatusOK)

        flusher, ok := w.(http.Flusher)
        if !ok {
            return fmt.Errorf("streaming not supported")
        }

        for i := range 10 {
            select {
            case <-ctx.Done():
                return nil  // client disconnected
            default:
            }
            fmt.Fprintf(w, "data: {\"event\": %d}\n\n", i)
            flusher.Flush()
            time.Sleep(time.Second)
        }
        return nil
    }
    return out, nil
}

zorya.Get(api, "/events", streamEvents)
```

## Chunked JSON stream

Stream a large result set as newline-delimited JSON (NDJSON):

```go
func streamUsers(ctx context.Context, _ *struct{}) (*EventStreamOutput, error) {
    out := &EventStreamOutput{}
    out.Body = func(w http.ResponseWriter) error {
        w.Header().Set("Content-Type", "application/x-ndjson")
        enc := json.NewEncoder(w)
        for _, u := range db.AllUsers() {
            if err := enc.Encode(u); err != nil {
                return err
            }
            w.(http.Flusher).Flush()
        }
        return nil
    }
    return out, nil
}
```

## Notes

- Streaming bodies bypass response transformers. Transformers only run for struct bodies.
- Streaming bodies bypass content negotiation. Set `Content-Type` explicitly in the function.
- The handler's returned error is used only if the body function itself has not started writing. Once `w.WriteHeader` is called, errors cannot change the status code.
- Check `ctx.Done()` inside long-running stream loops to detect client disconnects.
