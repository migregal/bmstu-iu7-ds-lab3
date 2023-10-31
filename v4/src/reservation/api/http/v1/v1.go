package v1

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/httpwrapper"
	"github.com/migregal/bmstu-iu7-ds-lab2/reservation/core/ports/reservations"
)

type Core interface {
	GetUserReservations(context.Context, string, string) ([]reservations.Reservation, error)
	AddReservation(context.Context, string, reservations.Reservation) (string, error)
	UpdateUserReservation(context.Context, string, string) error
}

type api struct {
	lg *slog.Logger

	core Core
}

func InitListener(mx *echo.Echo, lg *slog.Logger, core Core) error {
	gr := mx.Group("/api/v1")

	a := api{lg: lg, core: core}

	gr.POST("/reservations", httpwrapper.WrapRequest(a.lg, a.AddReservation))
	gr.GET("/reservations", httpwrapper.WrapRequest(a.lg, a.GetReservations))
	gr.PATCH("/reservations/:id", httpwrapper.WrapRequest(a.lg, a.UpdateReservation))

	return nil
}
