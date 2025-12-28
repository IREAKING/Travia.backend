package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	PexelsAPIURL = "https://api.pexels.com/v1/search"
	MaxRetries   = 3
	RetryDelay   = 1 * time.Second
)

type PexelsService struct {
	APIKey string
	Client *http.Client
}

type PexelsResponse struct {
	Photos []struct {
		Src struct {
			Large    string `json:"large"`
			Medium   string `json:"medium"`
			Original string `json:"original"`
		} `json:"src"`
		Alt string `json:"alt"`
	} `json:"photos"`
	TotalResults int `json:"total_results"`
	Page         int `json:"page"`
	PerPage      int `json:"per_page"`
}

type PexelsError struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewPexelsService(apiKey string) *PexelsService {
	return &PexelsService{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchImage searches for images on Pexels based on the destination name and province
func (p *PexelsService) SearchImage(destinationName, province string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("Pexels API key is not configured")
	}

	// Create search query based on destination and province
	query := p.buildSearchQuery(destinationName, province)

	// Try multiple search strategies if the first one fails
	searchStrategies := []string{
		query,
		fmt.Sprintf("%s city", destinationName),
		fmt.Sprintf("%s travel", destinationName),
		fmt.Sprintf("%s tourism", destinationName),
		destinationName,
	}

	for i, searchQuery := range searchStrategies {
		imageURL, err := p.performSearch(searchQuery)
		if err == nil && imageURL != "" {
			return imageURL, nil
		}

		// Log the attempt for debugging
		fmt.Printf("Search attempt %d for '%s' failed: %v\n", i+1, searchQuery, err)

		// Add delay between retries to respect rate limits
		if i < len(searchStrategies)-1 {
			time.Sleep(RetryDelay)
		}
	}

	return "", fmt.Errorf("no suitable image found for destination: %s", destinationName)
}

// buildSearchQuery creates an optimized search query for Pexels
func (p *PexelsService) buildSearchQuery(destinationName, province string) string {
	// Clean and format the destination name
	destination := strings.TrimSpace(destinationName)
	province = strings.TrimSpace(province)

	// If province is provided and different from destination, include it
	if province != "" && province != destination {
		return fmt.Sprintf("%s %s city", destination, province)
	}

	// Default search strategy
	return fmt.Sprintf("%s city travel", destination)
}

// performSearch executes a single search request to Pexels API
func (p *PexelsService) performSearch(query string) (string, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("per_page", "1")
	params.Set("orientation", "landscape") // Prefer landscape images for destinations
	params.Set("size", "large")            // Prefer large images

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", PexelsAPIURL, params.Encode()), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", p.APIKey)
	req.Header.Set("User-Agent", "Travia-Backend/1.0")

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var pexelsErr PexelsError
		if err := json.Unmarshal(body, &pexelsErr); err == nil {
			return "", fmt.Errorf("Pexels API error: %s (status: %d)", pexelsErr.Error, resp.StatusCode)
		}
		return "", fmt.Errorf("Pexels API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var data PexelsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(data.Photos) == 0 {
		return "", sql.ErrNoRows
	}

	// Return the best quality image available
	photo := data.Photos[0]
	if photo.Src.Original != "" {
		return photo.Src.Original, nil
	}
	if photo.Src.Large != "" {
		return photo.Src.Large, nil
	}
	if photo.Src.Medium != "" {
		return photo.Src.Medium, nil
	}

	return "", fmt.Errorf("no valid image URL found in response")
}

// ValidateAPIKey checks if the Pexels API key is valid
func (p *PexelsService) ValidateAPIKey() error {
	if p.APIKey == "" {
		return fmt.Errorf("Pexels API key is not configured")
	}

	req, err := http.NewRequest("GET", "https://api.pexels.com/v1/search?query=test&per_page=1", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", p.APIKey)
	req.Header.Set("User-Agent", "Travia-Backend/1.0")

	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid Pexels API key")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetRateLimitInfo returns information about API rate limits
func (p *PexelsService) GetRateLimitInfo() (remaining int, resetTime time.Time, err error) {
	req, err := http.NewRequest("GET", "https://api.pexels.com/v1/search?query=test&per_page=1", nil)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", p.APIKey)
	req.Header.Set("User-Agent", "Travia-Backend/1.0")

	resp, err := p.Client.Do(req)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Extract rate limit info from headers
	remainingStr := resp.Header.Get("X-Ratelimit-Remaining")
	resetStr := resp.Header.Get("X-Ratelimit-Reset")

	if remainingStr != "" {
		if remaining, err := fmt.Sscanf(remainingStr, "%d", &remaining); err == nil {
			// Parse reset time if available
			if resetStr != "" {
				if resetUnix, err := fmt.Sscanf(resetStr, "%d", &remaining); err == nil {
					resetTime = time.Unix(int64(resetUnix), 0)
				}
			}
			return remaining, resetTime, nil
		}
	}

	return 0, time.Time{}, fmt.Errorf("rate limit info not available")
}
