package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Event represents a typed SSE event
type Event struct {
	ID   string      `json:"id,omitempty"`
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data"`
}

// Writer handles SSE response streaming
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewWriter creates an SSE writer and sets required headers.
// Returns nil if streaming is not supported.
func NewWriter(w http.ResponseWriter) *Writer {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	return &Writer{w: w, flusher: flusher}
}

// Send writes a typed event to the stream
func (sw *Writer) Send(event Event) error {
	if event.ID != "" {
		fmt.Fprintf(sw.w, "id: %s\n", event.ID)
	}
	if event.Type != "" {
		fmt.Fprintf(sw.w, "event: %s\n", event.Type)
	}

	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	fmt.Fprintf(sw.w, "data: %s\n\n", data)
	sw.flusher.Flush()
	return nil
}

// SendKeepAlive sends a comment to keep the connection alive
func (sw *Writer) SendKeepAlive() {
	fmt.Fprint(sw.w, ": keepalive\n\n")
	sw.flusher.Flush()
}

// StreamWithKeepAlive runs until context is cancelled, sending events and keep-alives
func StreamWithKeepAlive(ctx context.Context, sw *Writer, interval time.Duration, eventCh <-chan Event) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("SSE: StreamWithKeepAlive started")
	for {
		select {
		case <-ctx.Done():
			log.Printf("SSE: StreamWithKeepAlive exiting - context done: %v", ctx.Err())
			return
		case event, ok := <-eventCh:
			if !ok {
				log.Printf("SSE: StreamWithKeepAlive exiting - channel closed")
				return
			}
			_ = sw.Send(event)
		case <-ticker.C:
			sw.SendKeepAlive()
		}
	}
}
