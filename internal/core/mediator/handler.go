package mediator

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Handler interface {
	Handle(ctx context.Context, cmd Message) (interface{}, error)
}

type Mediator struct {
	handlers map[string]Handler
}

type Option func(*Mediator) error

func New(options ...Option) (*Mediator, error) {
	mediator := &Mediator{
		handlers: make(map[string]Handler),
	}

	for _, option := range options {
		if err := option(mediator); err != nil {
			return nil, err
		}
	}

	return mediator, nil
}

func WithHandler(command Message, handler Handler) Option {
	return func(mediator *Mediator) error {
		mediator.handlers[command.Key()] = handler
		return nil
	}
}

func (m *Mediator) Send(ctx context.Context, command Message) (interface{}, error) {
	if handler, ok := m.handlers[command.Key()]; ok {
		result, err := handler.Handle(ctx, command)
		if err != nil {
			if _, ok := err.(*exceptions.NotFoundError); ok {
				metrics.NotFoundRequestError.With(prometheus.Labels{"method": command.Key()}).Inc()
			}

			if _, ok := err.(*exceptions.ConflictError); ok {
				metrics.ConflictRequestError.With(prometheus.Labels{"method": command.Key()}).Inc()
			}

			if _, ok := err.(*exceptions.ValidationError); ok {
				metrics.BadRequestError.With(prometheus.Labels{"method": command.Key()}).Inc()
			}

			if _, ok := err.(*exceptions.InternalError); ok {
				metrics.InternalError.With(prometheus.Labels{"method": command.Key()}).Inc()
			}

			panic(err)
		}

		return result, nil
	}
	return nil, errors.New("No handler registered for the command")
}
