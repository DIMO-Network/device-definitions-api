package commands

import (
	"context"
	"github.com/pkg/errors"
)

type CommandRequest interface {
	Key() string
}

type CommandResult interface {
}

type IHandler interface {
	Handle(ctx context.Context, cmd CommandRequest) (CommandResult, error)
}

type CustomMediator struct {
	handlers map[string]IHandler
}

type Option func(*CustomMediator) error

func New(options ...Option) (*CustomMediator, error) {
	mediator := &CustomMediator{
		handlers: make(map[string]IHandler),
	}

	for _, option := range options {
		if err := option(mediator); err != nil {
			return nil, err
		}
	}

	return mediator, nil
}

func WithHandler(command CommandRequest, handler IHandler) Option {
	return func(mediator *CustomMediator) error {
		mediator.handlers[command.Key()] = handler
		return nil
	}
}

func (m *CustomMediator) Send(ctx context.Context, command CommandRequest) (CommandResult, error) {
	if handler, ok := m.handlers[command.Key()]; ok {
		return handler.Handle(ctx, command)
	}
	return nil, errors.New("No handler registered for the command")
}

// todo: pending custom middleware
type Middleware interface {
	Execute(command CommandRequest, next HandlerFunc) (*CommandResult, error)
}

type PipelineExecution struct {
	mediator    *CustomMediator
	middlewares []Middleware
}
