package service

type Result struct {
	Output string
	// StatusCode should be 0 for undefined, or
	// the HTTP status code of the response.
	StatusCode int
}
