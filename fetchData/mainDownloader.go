package fetchdata

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

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
