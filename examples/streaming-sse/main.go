// streaming-sse demonstrates Server-Sent Events (SSE) using Zorya's streaming
// body pattern. The handler returns a function body that writes events directly
// to the response writer, checking context cancellation for clean shutdown.
//
// Run:
//
//	go run ./examples/streaming-sse
//
// Try:
//
//	curl -N http://localhost:8080/events
//
//	# Or in a browser: open http://localhost:8080/
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

// --- Input / output types ---

type EventStreamOutput struct {
	Body func(w http.ResponseWriter) error
}

// --- Handler ---

func streamEvents(ctx context.Context, _ *struct{}) (*EventStreamOutput, error) {
	out := &EventStreamOutput{}
	out.Body = func(w http.ResponseWriter) error {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("streaming not supported by this server")
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for i := range 10 {
			select {
			case <-ctx.Done():
				// Client disconnected
				return nil
			case t := <-ticker.C:
				event := fmt.Sprintf(`{"seq":%d,"time":"%s"}`, i, t.Format(time.RFC3339))
				fmt.Fprintf(w, "data: %s\n\n", event)
				flusher.Flush()
			}
		}

		// Signal end-of-stream
		fmt.Fprintf(w, "event: done\ndata: {}\n\n")
		flusher.Flush()
		return nil
	}
	return out, nil
}

// --- Simple HTML landing page ---

type PageOutput struct {
	Body func(w http.ResponseWriter) error
}

func landingPage(_ context.Context, _ *struct{}) (*PageOutput, error) {
	out := &PageOutput{}
	out.Body = func(w http.ResponseWriter) error {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body>
<h2>SSE Demo</h2>
<ul id="log"></ul>
<script>
const es = new EventSource("/events");
es.onmessage = e => {
    const li = document.createElement("li");
    li.textContent = e.data;
    document.getElementById("log").appendChild(li);
};
es.addEventListener("done", () => es.close());
</script></body></html>`)
		return nil
	}
	return out, nil
}

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(adapters.NewChi(router), zorya.WithConfig(&zorya.Config{
		OpenAPIPath: "/openapi.json",
		DocsPath:    "/docs",
	}))

	zorya.Get(api, "/", landingPage)
	zorya.Get(api, "/events", streamEvents)

	log.Println("Listening on :8080  —  open http://localhost:8080 in a browser")
	log.Fatal(http.ListenAndServe(":8080", router))
}
