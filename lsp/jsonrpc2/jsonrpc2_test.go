// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package jsonrpc2_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path"
	"reflect"
	"testing"

	"encoding/json"

	"github.com/a-h/templ/lsp/jsonrpc2"
)

const (
	methodNoArgs    = "no_args"
	methodOneString = "one_string"
	methodOneNumber = "one_number"
	methodJoin      = "join"
)

type callTest struct {
	method string
	params any
	expect any
}

var callTests = []callTest{
	{
		method: methodNoArgs,
		params: nil,
		expect: true,
	},
	{
		method: methodOneString,
		params: "fish",
		expect: "got:fish",
	},
	{
		method: methodOneNumber,
		params: 10,
		expect: "got:10",
	},
	{
		method: methodJoin,
		params: []string{"a", "b", "c"},
		expect: "a/b/c",
	},
	// TODO: expand the test cases
}

func (test *callTest) newResults() any {
	switch e := test.expect.(type) {
	case []any:
		var r []any
		for _, v := range e {
			r = append(r, reflect.New(reflect.TypeOf(v)).Interface())
		}
		return r

	case nil:
		return nil

	default:
		return reflect.New(reflect.TypeOf(test.expect)).Interface()
	}
}

func (test *callTest) verifyResults(t *testing.T, results any) {
	t.Helper()

	if results == nil {
		return
	}

	val := reflect.Indirect(reflect.ValueOf(results)).Interface()
	if !reflect.DeepEqual(val, test.expect) {
		t.Errorf("%v:Results are incorrect, got %+v expect %+v", test.method, val, test.expect)
	}
}

func TestRequest(t *testing.T) {
	ctx := context.Background()
	a, b, done := prepare(ctx, t)
	defer done()

	for _, test := range callTests {
		t.Run(test.method, func(t *testing.T) {
			results := test.newResults()
			if _, err := a.Call(ctx, test.method, test.params, results); err != nil {
				t.Fatalf("%v:Call failed: %v", test.method, err)
			}
			test.verifyResults(t, results)

			if _, err := b.Call(ctx, test.method, test.params, results); err != nil {
				t.Fatalf("%v:Call failed: %v", test.method, err)
			}
			test.verifyResults(t, results)
		})
	}
}

func prepare(ctx context.Context, t *testing.T) (a, b jsonrpc2.Conn, done func()) {
	t.Helper()

	// make a wait group that can be used to wait for the system to shut down
	aPipe, bPipe := net.Pipe()
	a = run(ctx, aPipe)
	b = run(ctx, bPipe)
	done = func() {
		a.Close()
		b.Close()
		<-a.Done()
		<-b.Done()
	}

	return a, b, done
}

func run(ctx context.Context, nc io.ReadWriteCloser) jsonrpc2.Conn {
	stream := jsonrpc2.NewStream(nc)
	conn := jsonrpc2.NewConn(stream)
	conn.Go(ctx, testHandler())

	return conn
}

func testHandler() jsonrpc2.Handler {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		switch req.Method() {
		case methodNoArgs:
			if len(req.Params()) > 0 {
				return reply(ctx, nil, fmt.Errorf("expected no params: %w", jsonrpc2.ErrInvalidParams))
			}
			return reply(ctx, true, nil)

		case methodOneString:
			var v string
			dec := json.NewDecoder(bytes.NewReader(req.Params()))
			if err := dec.Decode(&v); err != nil {
				return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
			}
			return reply(ctx, "got:"+v, nil)

		case methodOneNumber:
			var v int
			dec := json.NewDecoder(bytes.NewReader(req.Params()))
			if err := dec.Decode(&v); err != nil {
				return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
			}
			return reply(ctx, fmt.Sprintf("got:%d", v), nil)

		case methodJoin:
			var v []string
			dec := json.NewDecoder(bytes.NewReader(req.Params()))
			if err := dec.Decode(&v); err != nil {
				return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
			}
			return reply(ctx, path.Join(v...), nil)

		default:
			return jsonrpc2.MethodNotFoundHandler(ctx, reply, req)
		}
	}
}
