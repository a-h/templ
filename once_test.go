package templ_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

type renderLockTest struct {
	ctx      context.Context
	expected string
}

func TestOnceComponent(t *testing.T) {
	withHello := templ.WithChildren(context.Background(), templ.Raw("hello"))
	tests := []struct {
		name  string
		tests []renderLockTest
	}{
		{
			name: "renders nothing without children",
			tests: []renderLockTest{
				{
					ctx:      context.Background(),
					expected: "",
				},
			},
		},
		{
			name: "children are rendered",
			tests: []renderLockTest{
				{
					ctx:      templ.WithChildren(context.Background(), templ.Raw("hello")),
					expected: "hello",
				},
			},
		},
		{
			name: "children are rendered once per context",
			tests: []renderLockTest{
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
			tests: []renderLockTest{
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
			c := templ.MustNewRenderLock().Once()
			for i, test := range tt.tests {
				t.Run(fmt.Sprintf("render %d/%d", i+1, len(tt.tests)), func(t *testing.T) {
					html, err := templ.ToGoHTML(test.ctx, c)
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
	t.Run("different RenderLock objects have different state", func(t *testing.T) {
		ctx := templ.WithChildren(context.Background(), templ.Raw("hello"))
		c1 := templ.MustNewRenderLock().Once()
		c2 := templ.MustNewRenderLock().Once()
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
	t.Run("using the same provider multiple times limits rendering", func(t *testing.T) {
		ctx := templ.WithChildren(context.Background(), templ.Raw("hello"))
		provider := templ.MustNewRenderLock()
		c1 := provider.Once()
		c2 := provider.Once()
		var w strings.Builder
		if err := c1.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c2.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff := cmp.Diff("hello", w.String()); diff != "" {
			t.Errorf("unexpected diff:\n%v", diff)
		}
	})
	t.Run("using the same ID limits rendering", func(t *testing.T) {
		ctx := templ.WithChildren(context.Background(), templ.Raw("hello"))
		c1 := templ.MustNewRenderLock(templ.WithLockID("abc")).Once()
		c2 := templ.MustNewRenderLock(templ.WithLockID("abc")).Once()
		var w strings.Builder
		if err := c1.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := c2.Render(ctx, &w); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff := cmp.Diff("hello", w.String()); diff != "" {
			t.Errorf("unexpected diff:\n%v", diff)
		}
	})
}
