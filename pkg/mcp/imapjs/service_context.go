package imapjs

import (
	"context"

	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
)

type storedAccountResolverContextKey struct{}
type dialerContextKey struct{}

func withStoredAccountResolver(ctx context.Context, resolver smailnailjs.StoredAccountResolver) context.Context {
	return context.WithValue(ctx, storedAccountResolverContextKey{}, resolver)
}

func storedAccountResolverFromContext(ctx context.Context) (smailnailjs.StoredAccountResolver, bool) {
	if ctx == nil {
		return nil, false
	}
	resolver, ok := ctx.Value(storedAccountResolverContextKey{}).(smailnailjs.StoredAccountResolver)
	return resolver, ok
}

func withDialer(ctx context.Context, dialer smailnailjs.Dialer) context.Context {
	return context.WithValue(ctx, dialerContextKey{}, dialer)
}

func dialerFromContext(ctx context.Context) (smailnailjs.Dialer, bool) {
	if ctx == nil {
		return nil, false
	}
	dialer, ok := ctx.Value(dialerContextKey{}).(smailnailjs.Dialer)
	return dialer, ok
}
