package imapjs

import (
	"context"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
)

type resolvedIdentityContextKey struct{}

func withResolvedIdentity(ctx context.Context, resolved *identity.ResolvedIdentity) context.Context {
	return context.WithValue(ctx, resolvedIdentityContextKey{}, resolved)
}

func ResolvedIdentityFromContext(ctx context.Context) (*identity.ResolvedIdentity, bool) {
	if ctx == nil {
		return nil, false
	}
	resolved, ok := ctx.Value(resolvedIdentityContextKey{}).(*identity.ResolvedIdentity)
	return resolved, ok
}
