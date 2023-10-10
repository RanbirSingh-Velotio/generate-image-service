package main

import (
	"awesomeProject/generate-image-service/pkg/generateImage"
	generateImageHandler "awesomeProject/generate-image-service/pkg/generateImage/handler"
	generateImageService "awesomeProject/generate-image-service/pkg/generateImage/service"
	"awesomeProject/generate-image-service/utils/handlerutil"
	"fmt"
	"github.com/google/gops/agent"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
)

// startServer start server using grace
func startServer() {
	// gops for profiling
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Printf("Error")
	}

	defer agent.Close()

	// Start server

	serverPort := ":" + strconv.Itoa(8080)

	err := http.ListenAndServe(serverPort, nil)
	fmt.Printf("server started at Host:127.0.0.1:8080 ")
	if err != nil {
		log.Printf("unable to shutdown http server gracefully: %v\n", err)
	}
}

func handlerHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Success")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func initializegenerateImageService() {

	generateImageSrv := generateImageService.New()
	generateImage.Init(generateImageSrv)
	handler := generateImageHandler.InitHandler(generateImageSrv)
	handlerutil.Add(handler)
}

func main() {

	initializegenerateImageService()

	handlerutil.Start()

	http.HandleFunc("/", notFoundHandler)

	http.HandleFunc("/health", handlerHealthCheck)

	startServer()
}
