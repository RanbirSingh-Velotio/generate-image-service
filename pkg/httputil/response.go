package httputil

// StandardError is TopAds standard JSON HTTP Error.
type StandardError struct {
	Code   string      `json:"code"`
	Title  string      `json:"title"`
	Detail string      `json:"detail"`
	Object ErrorObject `json:"object"`
}

// ErrorObject holds any additional details of an error.
type ErrorObject struct {
	Text []string `json:"text"`
	Type int64    `json:"type"`
}
