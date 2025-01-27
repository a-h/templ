// SPDX-FileCopyrightText: 2020 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"context"
)

type ctxClientKey int

var ctxClient ctxClientKey = 0

// WithClient returns the context with Client value.
func WithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, ctxClient, client)
}

// ClientFromContext extracts Client from context.
func ClientFromContext(ctx context.Context) Client {
	client, ok := ctx.Value(ctxClient).(Client)
	if !ok {
		return nil
	}
	return client
}
