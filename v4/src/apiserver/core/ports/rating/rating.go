package rating

import (
	"context"
	"errors"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
	ErrUnavaliable       = errors.New("rating service unavailable")
)

type Config struct {
	Host     string
	Port     string
	MaxFails uint32 `mapstructure:"max_fails"`
}

type Client interface {
	GetUserRating(ctx context.Context, username string) (Rating, error)
	UpdateUserRating(ctx context.Context, username string, diff int) error
}
