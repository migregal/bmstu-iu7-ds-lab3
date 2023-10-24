package rating

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/migregal/bmstu-iu7-ds-lab2/apiserver/core/ports/rating"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/circuitbreaker"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/readiness"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/readiness/httpprober"
	v1 "github.com/migregal/bmstu-iu7-ds-lab2/rating/api/http/v1"
)

const probeKey = "http-rating-client"

var ErrInvalidStatusCode = errors.New("invalid status code")

type Client struct {
	lg *slog.Logger

	cb *circuitbreaker.Client

	conn *resty.Client
}

func New(lg *slog.Logger, cfg rating.Config, probe *readiness.Probe) (*Client, error) {
	client := resty.New().
		SetTransport(&http.Transport{
			MaxIdleConns:       10,               //nolint: gomnd
			IdleConnTimeout:    30 * time.Second, //nolint: gomnd
			DisableCompression: true,
		}).
		SetBaseURL(fmt.Sprintf("http://%s", net.JoinHostPort(cfg.Host, cfg.Port)))

	c := Client{
		lg:   lg,
		cb:   circuitbreaker.New(cfg.MaxFails),
		conn: client,
	}

	go httpprober.New(lg, client).Ping(probeKey, probe)

	return &c, nil
}

// nolint: dupl
func (c *Client) GetUserRating(
	ctx context.Context, username string,
) (rating.Rating, error) {
	if c.cb.Check("get_user_rating") {
		return rating.Rating{}, circuitbreaker.ErrSystemFails
	}

	res, err := c.getUserRating(ctx, username)
	if err != nil {
		c.cb.Inc("get_user_rating")

		return rating.Rating{}, err
	}

	c.cb.Release("get_user_rating")

	return res, nil
}

func (c *Client) getUserRating(
	_ context.Context, username string,
) (rating.Rating, error) {
	resp, err := c.conn.R().
		SetHeader("X-User-Name", username).
		SetResult(&v1.RatingResponse{}).
		Get("/api/v1/rating")
	if err != nil {
		return rating.Rating{}, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return rating.Rating{}, fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	data, _ := resp.Result().(*v1.RatingResponse)

	rating := rating.Rating(*data)

	return rating, nil
}

func (c *Client) UpdateUserRating(
	ctx context.Context, username string, diff int,
) error {
	if c.cb.Check("get_user_rating") {
		return circuitbreaker.ErrSystemFails
	}

	err := c.updateUserRating(ctx, username, diff)
	if err != nil {
		c.cb.Inc("get_user_rating")

		return err
	}

	c.cb.Release("get_user_rating")

	return nil
}

func (c *Client) updateUserRating(
	_ context.Context, username string, diff int,
) error {
	resp, err := c.conn.R().
		SetHeader("X-User-Name", username).
		SetQueryParam("diff", strconv.Itoa(diff)).
		SetResult(&v1.RatingResponse{}).
		Patch("/api/v1/rating")
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	return nil
}
