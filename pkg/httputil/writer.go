package httputil

import (
	"net/http"
)

type ResponseDecorator interface {
	Decorate(w http.ResponseWriter)
}

type ContentTypeDecorator string

func (d ContentTypeDecorator) Decorate(w http.ResponseWriter) {
	w.Header().Set("Content-Type", string(d))
}

func NewContentTypeDecorator(contentType string) ContentTypeDecorator {
	return ContentTypeDecorator(contentType)
}

type CORSDecorator struct {
	allowedOrigin string
}

func (d *CORSDecorator) Decorate(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", d.allowedOrigin)
}

func WriteResponse(w http.ResponseWriter, data []byte, status int, decorators ...ResponseDecorator) (int, error) {
	for _, decorator := range decorators {
		decorator.Decorate(w)
	}
	w.WriteHeader(status)
	return w.Write(data)
}
