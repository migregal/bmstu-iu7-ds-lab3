package reservation

import (
	"context"
	"errors"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
	ErrUnavaliable       = errors.New("reservation service unavailable")
)

type Config struct {
	Host     string
	Port     string
	MaxFails uint32 `mapstructure:"max_fails"`
}

type Client interface {
	GetUserReservations(ctx context.Context, username, status string) ([]Info, error)
	AddUserReservation(ctx context.Context, res Info) (string, error)
	SetUserReservationStatus(ctx context.Context, id, status string) error
}
