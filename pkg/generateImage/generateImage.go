package generateImage

import (
	"context"
	"time"
)

type GenerateImageRequestInput struct {
	Description       string `json:"description"`
	Colors            string `json:"colors"`
	Attributes        string `json:"attributes"`
	ChannelPlacements string `json:"channelPlacements"`
}

type Prediction struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Input     InputData `json:"input"`
	Logs      string    `json:"logs"`
	Error     string    `json:"error"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	URLs      struct {
		Cancel string `json:"cancel"`
		Get    string `json:"get"`
	} `json:"urls"`
}
type InputData struct {
	Prompt string `json:"prompt"`
}

type ReplicateResponse struct {
	ID    string `json:"id"`
	Input struct {
		Prompt string `json:"prompt"`
	} `json:"input"`
	Output []string `json:"output"`
	Status string   `json:"status"`
}
type GenerateImageResponse struct {
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl"`
	ID       string `json:"id"`
}

var RequestMap = make(map[string]GenerateImageResponse)

//go:generate mockgen -destination mockservice/mock_service.go -package mockservice github.com/RanbirSingh-Velotio/generateImage-service/pkg/generateImage Service
type Service interface {
	GenerateImageCreateRequest(ctx context.Context, requestInput GenerateImageRequestInput, id string) (GenerateImageResponse, error)
	GenerateImageGetRequest(ctx context.Context, reqIds string) GenerateImageResponse
}

var defaultService Service

func Init(svc Service) {
	defaultService = svc
}

func GetService() Service {
	return defaultService
}
