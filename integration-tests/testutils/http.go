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
	for i := 0; i < DefaultRetryAttempts; i++ {
		resp, err := http.Get(url)
		// We expect 200 OK if service is up, but connection refused if down
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(DefaultRetryInterval)
	}
	return fmt.Errorf("service at %s did not become available after %d attempts", url, DefaultRetryAttempts)
}
