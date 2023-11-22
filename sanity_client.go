package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	baseAPIURL = "https://%s.api.sanity.io/%s/data"
)

func RunQuery(query string) (string, error) {
	response, err := HTTPRequest("GET", "query", query)
	if err != nil {
		fmt.Println("Error querying document:", query)
		return "", err
	}
	// fmt.Println("Successfully queried document")
	return response, nil
}

func RunMutation(mutationData string) error {
	_, err := HTTPRequest("POST", "mutate", mutationData)
	if err != nil {
		fmt.Println("Error mutating document:", mutationData)
		return err
	}
	// fmt.Println("Successfully mutated document")
	return nil
}

// HTTPRequest performs a generic HTTP request and returns the response body as a string.
func HTTPRequest(method, path, data string) (string, error) {
	fullURL := fmt.Sprintf(baseAPIURL+"/%s/production", ProjectID, Version, path)

	var req *http.Request
	var err error

	if method == "GET" {
		fullURL += "?query=" + url.QueryEscape(data)
		req, err = http.NewRequest(method, fullURL, nil)
	} else if method == "POST" {
		req, err = http.NewRequest(method, fullURL, bytes.NewBuffer([]byte(data)))
	}

	if err != nil {
		return "", fmt.Errorf("error initializing request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}
