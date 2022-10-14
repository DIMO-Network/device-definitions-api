package exceptions

type ValidationError struct {
	Err error
}

func (r *ValidationError) Error() string {
	return r.Err.Error()
}

type ConflictError struct {
	Err error
}

func (r *ConflictError) Error() string {
	return r.Err.Error()
}

// NotFoundError will be translated by the behaviours to the correct Http or gRPC error status code
type NotFoundError struct {
	Err error
}

func (r *NotFoundError) Error() string {
	return r.Err.Error()
}

type InternalError struct {
	Err error
}

func (r *InternalError) Error() string {
	return r.Err.Error()
}
