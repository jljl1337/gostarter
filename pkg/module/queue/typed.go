package queue

import (
	"context"
	"encoding/json"
)

type Lane[T any] struct {
	name string
	q    *Service
}

func NewLane[T any](q *Service, name string) Lane[T] { return Lane[T]{name: name, q: q} }

func (l Lane[T]) Enqueue(ctx context.Context, v T) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return l.q.Enqueue(ctx, l.name, string(b))
}

func (l Lane[T]) Subscribe(h func(ctx context.Context, v T) error, opts ...LaneOption) {
	l.q.Subscribe(l.name, func(ctx context.Context, payload string) error {
		var v T
		if err := json.Unmarshal([]byte(payload), &v); err != nil {
			return err
		}
		return h(ctx, v)
	}, opts...)
}
