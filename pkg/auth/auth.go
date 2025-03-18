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

package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the authentication configuration
type Config struct {
	Token string `json:"token"`
}

// GetConfigDir returns the directory where the config file is stored
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "go-assessment")
	return configDir, nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// SaveToken saves the authentication token to the config file
func SaveToken(token string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	config := Config{
		Token: token,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// GetToken retrieves the authentication token from the config file
func GetToken() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not logged in, please run 'go-assessment login' first")
		}
		return "", fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("error unmarshaling config: %v", err)
	}

	if config.Token == "" {
		return "", fmt.Errorf("token not found, please run 'go-assessment login' first")
	}

	return config.Token, nil
}

// Login simulates authentication and stores the token
func Login(username, password string) error {
	// In a real application, this would make an API call to authenticate
	// For this assessment, we'll simulate authentication by generating a token
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	// Generate a simple token (in a real app, this would come from the server)
	token := fmt.Sprintf("simulated-token-%s-%s", username, password)

	return SaveToken(token)
}
