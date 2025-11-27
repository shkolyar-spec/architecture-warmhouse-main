package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TemperatureService handles fetching temperature data from external API
type TemperatureService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// TemperatureResponse represents the response from the temperature API
type TemperatureResponse struct {
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Location  string    `json:"location"`
}

// NewTemperatureService creates a new TemperatureService
func NewTemperatureService(baseURL string) *TemperatureService {
	return &TemperatureService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// GetTemperature fetches the temperature for a given location
func (s *TemperatureService) GetTemperature(location string) (*TemperatureResponse, error) {
	u, err := url.Parse(s.BaseURL + "/temperature")
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %w", err)
	}

	q := u.Query()
	q.Set("location", location)
	u.RawQuery = q.Encode()

	resp, err := s.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error calling temperature API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var temperatureResp TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&temperatureResp); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	return &temperatureResp, nil
}
