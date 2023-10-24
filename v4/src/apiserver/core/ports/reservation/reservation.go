package reservation

import "context"

type Config struct {
	Host     string
	Port     string
	MaxFails uint64 `mapstructure:"max_fails"`
}

type Client interface {
	GetUserReservations(ctx context.Context, username, status string) ([]Info, error)
	AddUserReservation(ctx context.Context, res Info) (string, error)
	SetUserReservationStatus(ctx context.Context, id, status string) error
}
