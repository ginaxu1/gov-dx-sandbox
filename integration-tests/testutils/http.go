package testutils

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// WaitForService waits for a service to become available by sending a request to the given URL.
// It retries for a specified number of times with a delay between attempts.
func WaitForService(url string) {
	for i := 0; i < 30; i++ {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte("{}")))
		// We expect 400 Bad Request or similar if service is up, but connection refused if down
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Service might not be ready, proceeding anyway...")
}
