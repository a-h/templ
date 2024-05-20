package templ_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

type onceTest struct {
	ctx      context.Context
	expected string
}

func TestOnceComponent(t *testing.T) {
	withHello := templ.WithChildren(context.Background(), templ.Raw("hello"))

	tests := []struct {
		name  string
		c     templ.OnceComponent
		tests []onceTest
	}{
		{
			name: "renders nothing without children",
			c:    templ.Once("id"),
			tests: []onceTest{
				{
					ctx:      context.Background(),
					expected: "",
				},
			},
		},
		{
			name: "children are rendered",
			c:    templ.Once("id"),
			tests: []onceTest{
				{
					ctx:      templ.WithChildren(context.Background(), templ.Raw("hello")),
					expected: "hello",
				},
			},
		},
		{
			name: "children are rendered once per context",
			c:    templ.Once("id"),
			tests: []onceTest{
				{
					ctx:      withHello,
					expected: "hello",
				},
				{
					ctx:      withHello,
					expected: "",
				},
			},
		},
		{
			name: "different contexts have different once state",
			c:    templ.Once("id"),
			tests: []onceTest{
				{
					ctx:      templ.WithChildren(context.Background(), templ.Raw("hello")),
					expected: "hello",
				},
				{
					ctx:      templ.WithChildren(context.Background(), templ.Raw("hello2")),
					expected: "hello2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, test := range tt.tests {
				t.Run(fmt.Sprintf("render %d/%d", i+1, len(tt.tests)), func(t *testing.T) {
					html, err := templ.ToGoHTML(test.ctx, tt.c)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if diff := cmp.Diff(test.expected, string(html)); diff != "" {
						t.Errorf("unexpected diff:\n%v", diff)
					}
				})
			}
		})
	}
	t.Run("different IDs have different once state", func(t *testing.T) {
		ctx := templ.WithChildren(context.Background(), templ.Raw("hello"))
		c1 := templ.Once("id1")
		c2 := templ.Once("id2")
		var w strings.Builder
		if err := c1.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c2.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff := cmp.Diff("hellohello", w.String()); diff != "" {
			t.Errorf("unexpected diff:\n%v", diff)
		}
	})
}
