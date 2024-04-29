package rest

import (
	"context"
	"fmt"
	"gitlab.com/grpasr/asonrythme/registry_svc/internal/config"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var server *http.Server
var serverWg sync.WaitGroup

func TestMain(m *testing.M) {

	startServer()

	// Run the tests
	result := m.Run()

	stopServer()

	os.Exit(result)
}

func startServer() {

	config.SetVarEnv()

	c := config.NewConfig()
	port := c.HTTPGetPort()
	address := c.HTTPGetAddress()

	rc := config.NewRestConfig()
	rc.RESTSetPathToStorage("../../../testsStorage")
	router := Handler(rc)

	server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", address, port),
		Handler: router,
	}

	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			// Handle other errors if needed
		}
	}()
}

func stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		// Handle the error if needed
	}

	serverWg.Wait()
}

func cltQuery(endpoint string) (statusCode int, respBody string, filesCount int, filesName []string, err error) {
	// func cltQuery(endpoint string) {
	resp, queryErr := http.Get(endpoint)
	if queryErr != nil {
		fmt.Printf("Error making GET request to serving server: %v\n", queryErr)
		return
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode

	// // // part to test response statusCode and contents
	// // if resp.StatusCode != http.StatusOK {
	// // 	// get the error
	// // }

	if resp.StatusCode == 200 {
		contentType := resp.Header.Get("Content-Type")
		// NOTE do that shit again................
		// if !isMultipart(contentType) {
		// 	fmt.Println("Expected multipart response, but received:", contentType)
		// 	return
		// }

		// Create a multipart reader
		multipartReader := multipart.NewReader(resp.Body, boundaryFromContentType(contentType))
		count := 0

		// Read each part
		for {
			part, err := multipartReader.NextPart()
			if err != nil {
				break
			}
			defer part.Close()

			// read the body only for the first file
			count++
			if count == 1 {
				body, readErr := io.ReadAll(part)
				if readErr != nil {
					// fmt.Printf("Error reading response body: %v\n", readErr)
					err = readErr
				}
				// fmt.Printf("Server response is: %v\n", string(body))
				respBody = string(body)
			}

			filesCount++
			filesName = append(filesName, part.FileName())
			// fmt.Printf("Part Header: %+v\n", part.Header)
		}
	} else {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			// fmt.Printf("Error reading response body: %v\n", readErr)
			err = readErr
			return
		}
		// fmt.Printf("Server response is: %v\n", string(body))
		respBody = string(body)

	}
	return
}

// Extract the boundary from the content type
func boundaryFromContentType(contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	return params["boundary"]
}
