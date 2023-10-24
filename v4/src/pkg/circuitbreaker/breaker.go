package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrSystemFails = errors.New("too many fails")

type fail struct {
	count uint64
	last  time.Time
}

type Client struct {
	maxFails uint64
	fails    map[string]fail
	mx       sync.RWMutex
}

func New(maxFails uint64) *Client {
	return &Client{
		maxFails: maxFails,
		fails:    make(map[string]fail),
	}
}

func (c *Client) Inc(method string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.fails[method] = fail{
		count: c.fails[method].count + 1,
		last:  time.Now(),
	}
}

func (c *Client) Check(method string) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	if c.fails[method].count >= c.maxFails {
		return true
	}

	return time.Now().Before(c.fails[method].last.Add(time.Minute))
}

func (c *Client) Release(method string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.fails[method] = fail{
		count: 0,
		last:  time.Time{},
	}
}
