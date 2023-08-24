package sse

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

func New() *Server {
	return &Server{
		requests: map[int64]chan Event{},
	}
}

type Server struct {
	counter  int64
	requests map[int64]chan Event
}

type Event struct {
	Type string
	Data string
}

func (s *Server) Send(eventType string, data string) {
	for _, f := range s.requests {
		f := f
		go func(f chan Event) {
			f <- Event{
				Type: eventType,
				Data: data,
			}
		}(f)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := atomic.AddInt64(&s.counter, 1)
	events := make(chan Event)
	s.requests[id] = events
	defer delete(s.requests, id)

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
			fmt.Println("Sending reload event...")
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
