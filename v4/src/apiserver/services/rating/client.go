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
	"github.com/sony/gobreaker"

	"github.com/migregal/bmstu-iu7-ds-lab2/apiserver/core/ports/rating"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/readiness"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/readiness/httpprober"
	"github.com/migregal/bmstu-iu7-ds-lab2/pkg/retryer"
	v1 "github.com/migregal/bmstu-iu7-ds-lab2/rating/api/http/v1"
)

const probeKey = "http-rating-client"

var ErrInvalidStatusCode = errors.New("invalid status code")

type ratingChange struct {
	username string
	diff     int
}

type Client struct {
	lg *slog.Logger

	cb      *gobreaker.CircuitBreaker
	retryer *retryer.Client[ratingChange]

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
		lg: lg,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "get_user_rating",
			Timeout:     time.Second,
			MaxRequests: cfg.MaxFails,
		}),
		retryer: retryer.New[ratingChange](),
		conn:    client,
	}

	go httpprober.New(lg, client).Ping(probeKey, probe)

	return &c, nil
}

func (c *Client) GetUserRating(
	ctx context.Context, username string,
) (rating.Rating, error) {
	data, err := c.cb.Execute(func() (any, error) {
		return c.getUserRating(ctx, username)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return rating.Rating{}, nil
		}

		return rating.Rating{}, fmt.Errorf("get rating: %w", err)
	}

	res, ok := data.(rating.Rating)
	if !ok {
		return rating.Rating{}, nil
	}

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
		return rating.Rating{}, fmt.Errorf("execute http request: %w", err)
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
	err := c.updateUserRating(ctx, username, diff)
	if err != nil {
		c.lg.Warn("failed to update rating", "err", err, "username", username)

		c.retryer.Append(ratingChange{
			username: username,
			diff:     diff,
		})
		c.retryer.Start(c.retryUpdate)
	}

	return nil
}

func (c *Client) retryUpdate(v ratingChange) {
	err := c.updateUserRating(context.Background(), v.username, v.diff)
	if err != nil {
		c.retryer.Append(v)
	}
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
		return fmt.Errorf("execute http request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%d: %w", resp.StatusCode(), ErrInvalidStatusCode)
	}

	return nil
}
