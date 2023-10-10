package service

import (
	"awesomeProject/generate-image-service/pkg/generateImage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type Service struct {
}

func New() *Service {
	service := &Service{}
	return service
}

const (
	replicateAPIToken = "dummy"
	version           = "1bfb924045802467cf8869d96b231a12e6aa994abfe37e337c63a4e49a8c6c41"
	url               = "https://api.replicate.com/v1/predictions"
)

const (
	responseStatusSucceeded = "succeeded"
)

var requestMap = make(map[string]generateImage.GenerateImageRequestInput)

func (s *Service) GenerateImageCreateRequest(ctx context.Context, brand generateImage.GenerateImageRequestInput, id string) (generateImage.GenerateImageResponse, error) {

	prompt := brand.Description + brand.Colors + brand.Attributes + brand.ChannelPlacements
	requestData := map[string]interface{}{
		"version": version,
		"input": map[string]string{
			"prompt": prompt,
		},
	}

	// Convert the request data to JSON
	requestDataJSON, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return generateImage.GenerateImageResponse{}, nil
	}

	// Create an HTTP POST request

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestDataJSON))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return generateImage.GenerateImageResponse{}, nil
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+replicateAPIToken)

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v\n", err)
		return generateImage.GenerateImageResponse{}, nil
	}
	defer resp.Body.Close()

	// Read the response body
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return generateImage.GenerateImageResponse{}, nil
	}

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("HTTP request failed with status: %d\n", resp.StatusCode)
		return generateImage.GenerateImageResponse{}, nil
	}

	var prediction generateImage.Prediction
	if err := json.Unmarshal(responseBytes, &prediction); err != nil {

		return generateImage.GenerateImageResponse{}, err
	}

	responseChan := make(chan generateImage.ReplicateResponse)
	var wg sync.WaitGroup

	// Add one to the WaitGroup for the Goroutine

	// Run the fetchImage function in a Goroutine
	go fetchImage(prediction.URLs.Get, replicateAPIToken, responseChan, &wg)

	defer close(responseChan)
	// Wait for the Goroutine to finish

	// Close the response channel when done

	var img image.Image
	// Process the response
	for {
		select {
		case response := <-responseChan:
			if response.Status == responseStatusSucceeded {
				img, err = downloadImage(response.Output[0])
				generateImage.RequestMap[id] = generateImage.GenerateImageResponse{
					Status:   "succeeded",
					ImageUrl: response.Output[0],
					ID:       id,
				}
				if err != nil {
					fmt.Printf("Error downloading image: %v\n", err)
					return generateImage.GenerateImageResponse{}, nil
				}

				fmt.Println("Image generation succeeded. Do further processing here.")
				// Create a new image context for rendering
				context := gg.NewContextForImage(img)

				// Create a new file to save the image
				outputFile, err := os.Create("output.png")
				if err != nil {
					fmt.Printf("Error creating output file: %v\n", err)
					return generateImage.GenerateImageResponse{}, nil
				}
				defer outputFile.Close()

				// Draw the image onto the context
				context.DrawImage(img, 0, 0)

				// Save the context as a PNG file
				if err := context.EncodePNG(outputFile); err != nil {
					fmt.Printf("Error encoding PNG: %v\n", err)
					return generateImage.GenerateImageResponse{}, nil
				}

				fmt.Println("Image saved as output.png")
				return generateImage.GenerateImageResponse{
					Status:   "succeeded",
					ImageUrl: response.Output[0],
				}, nil
			}
		}
	}
}

func downloadImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	img, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}
func fetchImage(imageUrl string, replicateAPIToken string, responseChan chan generateImage.ReplicateResponse, wg *sync.WaitGroup) {

	var response generateImage.ReplicateResponse
	req, err := http.NewRequest("POST", imageUrl, nil) // Replace requestBody with your actual request body
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return
	}
	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+replicateAPIToken)

	for {
		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending HTTP request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		responseBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return
		}

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("HTTP request failed with status: %d\n", resp.StatusCode)
			return
		}

		if err := json.Unmarshal(responseBytes, &response); err != nil {
			fmt.Println("Invalid JSON data:", err)
			responseChan <- response
			return
		}

		// Check if the status is "succeeded," and if so, send the response through the channel and exit
		if response.Status == responseStatusSucceeded {
			responseChan <- response
			return
		}

		// Sleep for a while before making the next request
		time.Sleep(1 * time.Second)
	}
}

func (s *Service) GenerateImageGetRequest(ctx context.Context, reqIds string) generateImage.GenerateImageResponse {

	return generateImage.RequestMap[reqIds]
}
