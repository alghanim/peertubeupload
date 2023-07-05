package httpclient

import (
	"net/http"
	"time"
)

// Create a new HTTP client with the specified settings
func New() *http.Client {
	transport := &http.Transport{
		MaxConnsPerHost:     10,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	client := &http.Client{
		Timeout:   time.Minute * 10,
		Transport: transport,
	}

	return client
}
