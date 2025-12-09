package testutils

import (
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultRetryAttempts = 30
	DefaultRetryInterval = 1 * time.Second
)

// WaitForService waits for a service to become available by sending a request to the given URL.
// It retries for a specified number of times with a delay between attempts.
func WaitForService(url string) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	for i := 0; i < DefaultRetryAttempts; i++ {
		resp, err := client.Get(url)
		if err != nil {
			// Connection error - service not available yet
			time.Sleep(DefaultRetryInterval)
			continue
		}

		// Ensure response body is closed
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		// Non-200 status - service might be starting
		time.Sleep(DefaultRetryInterval)
	}

	return fmt.Errorf("service at %s did not become available after %d attempts", url, DefaultRetryAttempts)
}
