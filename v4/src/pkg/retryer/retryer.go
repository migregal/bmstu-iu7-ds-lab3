package retryer

import (
	"slices"
	"sync"
	"time"
)

type Client[T any] struct {
	queue []T
	mx    sync.RWMutex

	start sync.Once
}

func New[T any]() *Client[T] {
	return &Client[T]{}
}

func (c *Client[T]) Append(v T) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.queue = append(c.queue, v)
}

func (c *Client[T]) Start(f func(T)) {
	try := func() {
		c.mx.Lock()

		if len(c.queue) == 0 {
			c.mx.Unlock()
			return
		}

		i := c.queue[0]
		c.queue = slices.Delete(c.queue, 0, 1)

		c.mx.Unlock()

		f(i)
	}

	c.start.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ticker.C:
				}

				try()
			}
		}()
	})

}
