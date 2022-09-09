package common

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type LoggingBehavior struct {
	log      *zerolog.Logger
	settings *config.Settings
}

func NewLoggingBehavior(log *zerolog.Logger, settings *config.Settings) LoggingBehavior {
	return LoggingBehavior{log: log, settings: settings}
}

func (p LoggingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
	p.log.Info().Msg(fmt.Sprintf("%s request logging : %v - %+v", p.settings.ServiceName, msg.Key(), msg))

	return next(ctx)
}

type ValidationBehavior struct {
	log      *zerolog.Logger
	settings *config.Settings
}

func NewValidationBehavior(log *zerolog.Logger, settings *config.Settings) ValidationBehavior {
	return ValidationBehavior{log: log, settings: settings}
}

// Process validation check for all requests going through mediator. Logs if validation fails.
func (p ValidationBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
	valErrors := Validate(msg)
	if valErrors != nil {
		// consider if reduce to Warn()
		p.log.Error().Msg(fmt.Sprintf("%s validation error : %v - %+v", p.settings.ServiceName, msg.Key(), msg))
		panic(exceptions.ValidationError{
			Err: errors.New(valErrors[0].Field),
		})
	}
	return next(ctx)
}

type ErrorHandlingBehavior struct {
	prometheusMetricService metrics.PrometheusMetricService
	log                     *zerolog.Logger
	settings                *config.Settings
}

func NewErrorHandlingBehavior(prometheusMetricService metrics.PrometheusMetricService, log *zerolog.Logger, settings *config.Settings) ErrorHandlingBehavior {
	return ErrorHandlingBehavior{prometheusMetricService: prometheusMetricService, log: log, settings: settings}
}

// Process checks for errors in the pipeline to increment metrics and log in standard fashion
func (p ErrorHandlingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
	r, err := next(ctx)
	if err != nil {
		// increment error metric
		p.prometheusMetricService.InternalError()
		// msg.Key contains the property names, and msg contains the property values that were passed into the function to execute.
		// this automatically logs any incoming properties for easy debugging. An improvement here could be to use reflection to map out the properties to the log context.
		p.log.Error().
			Err(err).
			Msg(fmt.Sprintf("%s request error : %v - %+v", p.settings.ServiceName, msg.Key(), msg))
		//return nil, err // if just return error does not cut mediator pipeline and will continue normal execution, must panic for mediator to stop pipeline and go to error path
		panic(err)
	}
	//reflect.TypeOf(next).Name() to get name of the method
	// if no error, increment overall success metric
	p.prometheusMetricService.Success()

	return r, nil
}
