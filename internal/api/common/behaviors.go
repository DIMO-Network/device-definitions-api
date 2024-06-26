package common

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/DIMO-Network/device-definitions-api/internal/config"
//	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
//	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
//	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/rs/zerolog"
//)
//
//type LoggingBehavior struct {
//	log      *zerolog.logger
//	settings *config.settings
//}
//
//func NewLoggingBehavior(log *zerolog.logger, settings *config.settings) LoggingBehavior {
//	return LoggingBehavior{log: log, settings: settings}
//}
//
//func (p LoggingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
//	p.log.Debug().Msg(fmt.Sprintf("%s request logging : %v - %+v", p.settings.ServiceName, msg.key(), msg))
//
//	return next(ctx)
//}
//
//type ValidationBehavior struct {
//	log      *zerolog.logger
//	settings *config.settings
//}
//
//func NewValidationBehavior(log *zerolog.logger, settings *config.settings) ValidationBehavior {
//	return ValidationBehavior{log: log, settings: settings}
//}
//
//// Process validation check for all requests going through mediator. Logs if validation fails.
//func (p ValidationBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
//	valErrors := Validate(msg)
//	if valErrors != nil {
//		// consider if reduce to Warn()
//		p.log.Error().Msg(fmt.Sprintf("%s validation error : %v - %+v", p.settings.ServiceName, msg.key(), msg))
//		err := exceptions.ValidationError{Err: fmt.Errorf("The field %s is required", valErrors[0].Field)}
//
//		metrics.BadRequestError.With(prometheus.Labels{"method": msg.key()}).Inc()
//		panic(&err)
//	}
//	return next(ctx)
//}
//
//type ErrorHandlingBehavior struct {
//	log      *zerolog.logger
//	settings *config.settings
//}
//
//func NewErrorHandlingBehavior(log *zerolog.logger, settings *config.settings) ErrorHandlingBehavior {
//	return ErrorHandlingBehavior{log: log, settings: settings}
//}
//
//// Process checks for errors in the pipeline to increment metrics and log in standard fashion
//func (p ErrorHandlingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {
//	r, err := next(ctx)
//	if err != nil {
//
//		// msg.key contains the property names, and msg contains the property values that were passed into the function to execute.
//		// this automatically logs any incoming properties for easy debugging. An improvement here could be to use reflection to map out the properties to the log context.
//		p.log.Error().
//			Err(err).
//			Msg(fmt.Sprintf("%s request error : %v - %+v", p.settings.ServiceName, msg.key(), msg))
//
//		// if just return error does not cut mediator pipeline and will continue normal execution, must panic for mediator to stop pipeline and go to error path
//		// increment error metric
//		if _, ok := err.(*exceptions.ConflictError); ok {
//			metrics.ConflictRequestError.With(prometheus.Labels{"method": msg.key()}).Inc()
//		}
//
//		if _, ok := err.(*exceptions.NotFoundError); ok {
//			metrics.NotFoundRequestError.With(prometheus.Labels{"method": msg.key()}).Inc()
//		}
//
//		metrics.InternalError.With(prometheus.Labels{"method": msg.key()}).Inc()
//		panic(err)
//	}
//	//reflect.TypeOf(next).Name() to get name of the method
//	// if no error, increment overall success metric
//	metrics.Success.With(prometheus.Labels{"method": msg.key()}).Inc()
//
//	return r, nil
//}
