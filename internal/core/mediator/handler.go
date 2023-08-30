package mediator

import (
	"context"
	"github.com/pkg/errors"
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
		return handler.Handle(ctx, command)
	}
	return nil, errors.New("No handler registered for the command")
}
