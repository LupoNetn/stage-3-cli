package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RequestOptions defines the parameters for MakeRequest
type RequestOptions struct {
	Method  string
	URL     string
	Body    any
	Token   string
	Timeout time.Duration
}

// MakeRequest is a robust, general-purpose HTTP request function for JSON responses
func MakeRequest(opts RequestOptions) (map[string]any, error) {
	resp, err := doRequestWithRetry(opts, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// DownloadFile is for requests that return raw file data
func DownloadFile(opts RequestOptions) ([]byte, string, error) {
	resp, err := doRequestWithRetry(opts, true)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	filename := "export.csv" // Default
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		// Simple filename extraction
		if parts := strings.Split(cd, "filename="); len(parts) > 1 {
			filename = strings.Trim(parts[1], "\"")
		}
	}

	return data, filename, nil
}

func doRequestWithRetry(opts RequestOptions, retryOn401 bool) (*http.Response, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 15 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	var bodyReader io.Reader
	if opts.Body != nil {
		bodyBytes, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/csv")
	req.Header.Set("X-API-Version", "1")

	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == 401 && retryOn401 {
		resp.Body.Close() // Close previous body
		// Attempt Token Refresh
		cfg, err := LoadConfig()
		if err == nil && cfg.RefreshToken != "" {
			newTokens, refreshErr := refreshTokens(cfg.RefreshToken)
			if refreshErr == nil {
				// Update Config
				at, ok1 := newTokens["access_token"].(string)
				rt, ok2 := newTokens["refresh_token"].(string)
				if ok1 && ok2 {
					cfg.AccessToken = at
					cfg.RefreshToken = rt
					SaveConfig(*cfg)

					// Retry original request with new token
					opts.Token = cfg.AccessToken
					return doRequestWithRetry(opts, false)
				}
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var errData map[string]any
		json.NewDecoder(resp.Body).Decode(&errData)
		
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("session expired or invalid. Please run 'insighta login' again")
		}

		if msg, ok := errData["message"].(string); ok {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return resp, nil
}

func refreshTokens(refreshToken string) (map[string]any, error) {
	backendBaseUrl := os.Getenv("DEV_BACKEND_BASE_URL")
	if backendBaseUrl == "" {
		backendBaseUrl = "https://stage-3-backend-azure.vercel.app"
	}

	body := map[string]string{"refresh_token": refreshToken}
	bodyBytes, _ := json.Marshal(body)

	resp, err := http.Post(backendBaseUrl+"/auth/refresh", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to refresh token")
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	// The refresh endpoint returns a flat map
	return result, nil
}
