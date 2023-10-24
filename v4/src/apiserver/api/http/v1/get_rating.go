package v1

import (
	"errors"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RatingRequest struct {
	AuthedRequest `valid:"optional"`
}

type RatingResponse struct {
	Rating
}

func (a *api) GetRating(c echo.Context, req RatingRequest) error {
	data, err := a.core.GetUserRating(c.Request().Context(), req.Username)
	if err != nil {
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			return c.NoContent(http.StatusServiceUnavailable)
		}

		return c.NoContent(http.StatusInternalServerError)
	}

	resp := RatingResponse{
		Rating: Rating(data),
	}

	return c.JSON(http.StatusOK, &resp)
}
