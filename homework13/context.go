package rpc

import "context"

type OnewayKey struct {
}

func CtxWithOneway(ctx context.Context) context.Context {
	return context.WithValue(ctx, OnewayKey{}, true)
}

func isOneway(ctx context.Context) bool {
	val := ctx.Value(OnewayKey{})
	oneway, ok := val.(bool)
	return oneway && ok
}
