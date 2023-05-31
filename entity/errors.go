package entity

const (
	// MethodNotAllowed is a format string for the errors related to the respective http status
	MethodNotAllowed = "method %s not allowed"
	// UnableToReadTheBody is a string for the errors when we are unable to read the request body
	UnableToReadTheBody = "unable to read the body"
	// RequestBodyIsTooBig is a string for the errors when the request body is too big
	RequestBodyIsTooBig = "request body is too big"
	// BadRequest is a format string for the errors request is malformed
	BadRequest = "bad request, err: %s"
	// FailedToProcessTheRequest is a format string for the errors when the request processing failed
	FailedToProcessTheRequest = "failed to process the request, err: %s"
	// FailedToProcessTheResponse is a format string for the errors when the response processing failed
	FailedToProcessTheResponse = "failed to process the response, err: %s"
	// FailedToWriteTheResponse is a format string for the errors when we failed to write back the http response
	FailedToWriteTheResponse = "failed to write the response, err: %s"
)
