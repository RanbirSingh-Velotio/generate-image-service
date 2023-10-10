package handler

import (
	"awesomeProject/generate-image-service/pkg/generateImage"
	"awesomeProject/generate-image-service/pkg/httputil"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Handler struct {
	service generateImage.Service
}

var (
	errBadRequest     = errors.New("BAD_REQUEST")
	errRequestTimeOut = errors.New("REQUEST_TIMEOUT")
)

func InitHandler(service generateImage.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetIdentity returns handler identity
func (h *Handler) GetIdentity() string {
	return "generateImage-v1"
}

// Start will start all http handlers
func (h *Handler) Start() error {
	http.Handle("/v1/generateImage", TraceMiddleware(h))
	return nil
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) errorResponse(w http.ResponseWriter, code int) {

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleRequest(w, r)
	case http.MethodGet:
		h.HandleGetRequest(w, r)
	default:
		// Return error immediately if the request method is incorrect
		h.errorResponse(w, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {

	var input generateImage.GenerateImageRequestInput

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// handle error
		return
	}
	hash := sha1.Sum(body)
	uniqueID := fmt.Sprintf("%x", hash)
	err = json.Unmarshal(body, &input)
	var response generateImage.GenerateImageResponse
	if data, ok := generateImage.RequestMap[uniqueID]; ok {
		response = generateImage.GenerateImageResponse{
			Status:   "succeeded",
			ImageUrl: data.ImageUrl,
			ID:       uniqueID,
		}
	} else {
		generateImage.RequestMap[uniqueID] = generateImage.GenerateImageResponse{
			Status:   "",
			ImageUrl: "",
			ID:       uniqueID,
		}
		if err != nil {
			// handle error
			return
		}

		go func() {
			_, _ = h.service.GenerateImageCreateRequest(r.Context(), input, uniqueID)
		}()

		response = generateImage.GenerateImageResponse{
			Status:   "processing",
			ImageUrl: "",
			ID:       uniqueID,
		}
	}

	jsonResponse, _ := json.Marshal(response)

	w.Write(jsonResponse)
}

func (h *Handler) parseQueryParam(ctx context.Context, r *http.Request) (string, error) {

	idsParam := r.URL.Query().Get("ids")

	if idsParam == "" {
		return "", nil
	}

	return idsParam, nil
}

func (h *Handler) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	errChan := make(chan error, 1)
	var response generateImage.GenerateImageResponse
	defer func(start time.Time) {
		jsonResponse, _ := json.Marshal(response)
		_, err := httputil.WriteResponse(w, jsonResponse, http.StatusOK, httputil.NewContentTypeDecorator("application/json"))
		if err != nil {
			return
		}
	}(time.Now())
	go func(ctx context.Context) {
		var reqIds string
		reqIds, err = h.parseQueryParam(ctx, r)
		if err != nil {
			errChan <- err
			return
		}

		response = h.service.GenerateImageGetRequest(ctx, reqIds)
		errChan <- nil
	}(ctx)

	select {
	case <-ctx.Done():
		return
	case err = <-errChan:
		return
	}
}
