package fetchdata

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func DownloadManySECFiles(downloadLinks []string, filePaths []string) error {
	if len(downloadLinks) != len(filePaths) {
		return errors.New("the length of downloadLinks and filePaths must match")
	}

	// Create a ticker that emits every second.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Use a wait group to wait for all the goroutines to finish.
	var wg sync.WaitGroup

	// Error channel to collect errors from goroutines.
	errorsCh := make(chan error, len(downloadLinks))

	// This channel will block goroutines once it's filled until a tick allows more to proceed.
	semaphore := make(chan struct{}, 10)

	for i, link := range downloadLinks {
		// Increment the WaitGroup counter.
		wg.Add(1)

		// Start a goroutine for each download.
		go func(link, path string) {
			defer wg.Done()

			// Wait for the signal to start or for the ticker to allow another request.
			<-ticker.C
			semaphore <- struct{}{} // Acquire a token.

			// Perform the download.
			err := DownloadOneSECFile(link, path)
			if err != nil {
				errorsCh <- err
			}

			<-semaphore // Release the token.
		}(link, filePaths[i])
	}

	// Close the errors channel when all goroutines are done.
	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	// Collect errors from the error channel.
	var allErrors error
	for err := range errorsCh {
		if allErrors == nil {
			allErrors = err
		} else {
			allErrors = fmt.Errorf("%v; %v", allErrors, err)
		}
	}

	return allErrors
}

func DownloadOneSECFile(downloadLink, filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Get the environment variables
	userAgent := os.Getenv("USER_AGENT")
	companyName := os.Getenv("COMPANY_NAME")
	email := os.Getenv("EMAIL")

	// Create a request with the User-Agent header and query parameters
	client := &http.Client{}
	req, err := http.NewRequest("GET", downloadLink, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	// Set query parameters
	queryParams := req.URL.Query()
	queryParams.Set("company", companyName)
	queryParams.Set("email", email)
	req.URL.RawQuery = queryParams.Encode()

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the data to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Print the success message
	fmt.Printf("Download complete: %s\n", filePath)
	return nil
}
