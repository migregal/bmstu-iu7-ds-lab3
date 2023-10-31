package redis

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/adjust/rmq/v5"
)

type Config struct {
	Tag  string
	Addr string
}
type Client[T any] struct {
	queue rmq.Queue

	start sync.Once
}

func New[T any]() (*Client[T], error) {
	connection, err := rmq.OpenConnection(
		"my service", "tcp", "localhost:6379", 1, make(chan<- error, 100), //nolint: gomnd
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	queue, err := connection.OpenQueue("tasks")
	if err != nil {
		return nil, fmt.Errorf("open queue: %w", err)
	}

	return &Client[T]{queue: queue}, nil
}

func (c *Client[T]) Append(v T) error {
	taskBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("format data: %w", err)
	}

	err = c.queue.Publish(string(taskBytes))
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func (c *Client[T]) Start(f func(T) error) error {
	c.start.Do(func() {
		err := c.queue.StartConsuming(10, time.Second) //nolint: gomnd
		if err != nil {
			slog.Error("failed to start consuming", "err", err)

			return
		}

		taskConsumer := &TaskConsumer[T]{f: f}
		_, err = c.queue.AddConsumer("task-consumer", taskConsumer)
		if err != nil {
			slog.Error("failed to add consumer", "err", err)
		}
	})

	return nil
}

type TaskConsumer[T any] struct {
	f func(T) error
}

func (consumer *TaskConsumer[T]) Consume(delivery rmq.Delivery) {
	var task T
	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		slog.Error("failed to unmarshal task", "err", err)

		if err := delivery.Reject(); err != nil {
			slog.Error("failed to reject task", "err", err)
		}

		return
	}

	if err := consumer.f(task); err != nil {
		slog.Error("retry", "err", err)
	}

	if err := delivery.Ack(); err != nil {
		slog.Error("failed to ack task", "err", err)
	}
}
