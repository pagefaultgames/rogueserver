package errors

type HttpError struct {
	Code    int
	Message string
}

func NewHttpError(code int, message string) *HttpError {
	return &HttpError{Code: code, Message: message}
}

func (h HttpError) Error() string {
	return h.Message
}
