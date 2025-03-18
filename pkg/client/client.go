/*
Copyright 2025 LuisEspinoza.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an API client for interacting with the backend
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// DeployRequest represents the request body for deploying an application
type DeployRequest struct {
	Image       string `json:"image"`
	Name        string `json:"name"`
	MemoryLimit string `json:"memoryLimit"`
	MinReplicas int32  `json:"minReplicas"`
	MaxReplicas int32  `json:"maxReplicas"`
}

// StatusResponse represents the response from the status endpoint
type StatusResponse struct {
	Status   string `json:"status"`
	Replicas int32  `json:"replicas"`
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Token: token,
	}
}

// Deploy sends a request to deploy an application
func (c *Client) Deploy(req DeployRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/deploy", c.BaseURL), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deployment failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetStatus retrieves the status of a deployment
func (c *Client) GetStatus(name string) (*StatusResponse, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/status/%s", c.BaseURL, name), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var statusResp StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &statusResp, nil
}

// DestroyDeployment deletes a deployment
func (c *Client) DestroyDeployment(name string) error {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", c.BaseURL, name), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("destroy request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
