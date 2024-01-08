package fetchdata

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// func DownloadManySECFiles(downloadLinks []string, filePaths []string) error {
// 	if len(downloadLinks) != len(filePaths) {
// 		return errors.New("the length of downloadLinks and filePaths must match")
// 	}

// 	// Create a ticker that emits every second.
// 	ticker := time.NewTicker(1 * time.Second)
// 	defer ticker.Stop()

// 	// Use a wait group to wait for all the goroutines to finish.
// 	var wg sync.WaitGroup

// 	// Error channel to collect errors from goroutines.
// 	errorsCh := make(chan error, len(downloadLinks))

// 	// This channel will block goroutines once it's filled until a tick allows more to proceed.
// 	semaphore := make(chan struct{}, 10)

// 	for i, link := range downloadLinks {
// 		// Increment the WaitGroup counter.
// 		wg.Add(1)

// 		// Start a goroutine for each download.
// 		go func(link, path string) {
// 			defer wg.Done()

// 			// Wait for the signal to start or for the ticker to allow another request.
// 			<-ticker.C
// 			semaphore <- struct{}{} // Acquire a token.

// 			// Perform the download.
// 			err := DownloadOneSECFile(link, path)
// 			if err != nil {
// 				errorsCh <- err
// 			}

// 			<-semaphore // Release the token.
// 		}(link, filePaths[i])
// 	}

// 	// Close the errors channel when all goroutines are done.
// 	go func() {
// 		wg.Wait()
// 		close(errorsCh)
// 	}()

// 	// Collect errors from the error channel.
// 	var allErrors error
// 	for err := range errorsCh {
// 		if allErrors == nil {
// 			allErrors = err
// 		} else {
// 			allErrors = fmt.Errorf("%v; %v", allErrors, err)
// 		}
// 	}

// 	return allErrors
// }

func DownloadManySECFiles(downloadLinks []string, filePaths []string) error {
	if len(downloadLinks) != len(filePaths) {
		return errors.New("the length of downloadLinks and filePaths must match")
	}

	// A channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, 10)

	var wg sync.WaitGroup
	for i := 0; i < len(downloadLinks); i++ {
		wg.Add(1)
		go func(link, path string) {
			defer wg.Done()
			// Acquire a token
			semaphore <- struct{}{}
			// Ensure the token is released after this function finishes
			defer func() { <-semaphore }()
			// Call the download function
			if err := DownloadOneSECFile(link, path); err != nil {
				fmt.Println("Error downloading file:", err)
			}
		}(downloadLinks[i], filePaths[i])
	}

	// Wait for all downloads to finish
	wg.Wait()
	return nil
}

func DownloadOneSECFile(downloadLink string, filePath string) error {
	userAgent := os.Getenv("USER_AGENT")
	companyName := os.Getenv("COMPANY_NAME")
	email := os.Getenv("EMAIL")

	// Create a new request using the downloadLink
	req, err := http.NewRequest("GET", downloadLink, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", fmt.Sprintf("%s - %s (mailto:%s)", userAgent, companyName, email))

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	// Create the necessary parent directories for the file
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Open the file to write the response body into
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("File successfully downloaded and saved to: %s\n", filePath)
	return nil
}
