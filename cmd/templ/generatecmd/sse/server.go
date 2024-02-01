package sse

import (
	_ "embed"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func New() *Handler {
	return &Handler{
		m:        new(sync.Mutex),
		requests: map[int64]chan event{},
	}
}

type Handler struct {
	m        *sync.Mutex
	counter  int64
	requests map[int64]chan event
}

type event struct {
	Type string
	Data string
}

// Send an event to all connected clients.
func (s *Handler) Send(eventType string, data string) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, f := range s.requests {
		f := f
		go func(f chan event) {
			f <- event{
				Type: eventType,
				Data: data,
			}
		}(f)
	}
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := atomic.AddInt64(&s.counter, 1)
	s.m.Lock()
	events := make(chan event)
	s.requests[id] = events
	s.m.Unlock()
	defer func() {
		s.m.Lock()
		defer s.m.Unlock()
		delete(s.requests, id)
		close(events)
	}()

	timer := time.NewTimer(0)
loop:
	for {
		select {
		case <-timer.C:
			if _, err := fmt.Fprintf(w, "event: message\ndata: ping\n\n"); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			timer.Reset(time.Second * 5)
		case e := <-events:
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, e.Data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case <-r.Context().Done():
			break loop
		}
		w.(http.Flusher).Flush()
	}
}
