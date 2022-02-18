package kenall

import "fmt"

var (
	// ErrInvalidArgument is an error value that will be returned if the value of the argument is invalid.
	ErrInvalidArgument = fmt.Errorf("kenall: invalid argument")
	// ErrUnauthorized is an error value that will be returned if the authorization token is invalid.
	ErrUnauthorized = fmt.Errorf("kenall: 401 unauthorized error")
	// ErrPaymentRequired is an error value that will be returned if the payment for your kenall account is overdue.
	ErrPaymentRequired = fmt.Errorf("kenall: 402 payment required error")
	// ErrForbidden is an error value that will be returned when the resource does not have access privileges.
	ErrForbidden = fmt.Errorf("kenall: 403 forbidden error")
	// ErrNotFound is an error value that will be returned when there is no resource to be retrieved.
	ErrNotFound = fmt.Errorf("kenall: 404 not found error")
	// ErrMethodNotAllowed is an error value that will be returned when the request calls a method that is not allowed.
	ErrMethodNotAllowed = fmt.Errorf("kenall: 405 method not allowed error")
	// ErrInternalServerError is an error value that will be returned when some error occurs in the kenall service.
	ErrInternalServerError = fmt.Errorf("kenall: 500 internal server error")
	// ErrTimeout is an error value that will be returned when the request is timeout.
	ErrTimeout = func(err error) error { return fmt.Errorf("kenall: request timeout: %w", err) }
)
