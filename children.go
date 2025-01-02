package templ

import "context"

type contextChildrenType int

const contextChildrenKey = contextChildrenType(1)

type contextChildrenValue struct {
	children Component
}

func WithChildren(ctx context.Context, children Component) context.Context {
	ctx = InitializeContext(ctx)
	if children == nil {
		return ctx
	}
	return context.WithValue(ctx, contextChildrenKey, &contextChildrenValue{children: children})
}

func GetChildren(ctx context.Context) Component {
	if ctx == nil {
		return NopComponent
	}
	v, ok := ctx.Value(contextChildrenKey).(*contextChildrenValue)
	if !ok || v.children == nil {
		return NopComponent
	}
	return v.children
}

func ClearChildren(ctx context.Context) context.Context {
	if ctx == nil || ctx.Value(contextChildrenKey) == nil {
		return ctx
	}
	return context.WithValue(ctx, contextChildrenKey, nil)
}
