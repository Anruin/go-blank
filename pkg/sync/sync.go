package sync

import (
	"context"
	"github.com/anruin/go-blank/pkg/names"
	"golang.org/x/sync/errgroup"
)

func Initialize(ctx context.Context) (context.Context, error) {
	g := new(errgroup.Group)

	// Add the error group to the context.
	ctx = context.WithValue(ctx, names.CtxErrGroup, g)

	return ctx, nil
}
