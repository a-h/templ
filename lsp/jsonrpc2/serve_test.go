// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/a-h/templ/lsp/jsonrpc2"
)

func TestIdleTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = ln.Close()
	}()

	connect := func() net.Conn {
		conn, err := net.DialTimeout("tcp", ln.Addr().String(), 5*time.Second)
		if err != nil {
			panic(err)
		}
		return conn
	}

	server := jsonrpc2.HandlerServer(jsonrpc2.MethodNotFoundHandler)
	var (
		runErr error
		wg     sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runErr = jsonrpc2.Serve(ctx, ln, server, 100*time.Millisecond)
	}()

	// Exercise some connection/disconnection patterns, and then assert that when
	// our timer fires, the server exits.
	conn1 := connect()
	conn2 := connect()
	_ = conn1.Close()
	_ = conn2.Close()
	conn3 := connect()
	_ = conn3.Close()

	wg.Wait()

	if !errors.Is(runErr, jsonrpc2.ErrIdleTimeout) {
		t.Errorf("run() returned error %v, want %v", runErr, jsonrpc2.ErrIdleTimeout)
	}
}
