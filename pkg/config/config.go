package config

import "context"

type Config interface {
	Initialize(ctx context.Context) (context.Context, error)
}
