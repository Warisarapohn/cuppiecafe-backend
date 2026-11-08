package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type FoursquareClient struct {
	APIKey string
	Client *http.Client
}

type FoursquarePlace struct {
	FSQID     string  `json:"fsq_id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Locality  string  `json:"locality"`
	Region    string  `json:"region"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Category  string  `json:"category"`
	ImageURL  string  `json:"image_url"`
}

type foursquareSearchResponse struct {
	Results []struct {
		FSQPlaceID string `json:"fsq_place_id"`
		Name       string `json:"name"`

		Location struct {
			Address  string `json:"address"`
			Locality string `json:"locality"`
			Region   string `json:"region"`
		} `json:"location"`

		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`

		Categories []struct {
			Name string `json:"name"`
		} `json:"categories"`
	} `json:"results"`
}

func NewFoursquareClient() *FoursquareClient {
	return &FoursquareClient{
		APIKey: os.Getenv("FOURSQUARE_API_KEY"),
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (f *FoursquareClient) SearchCafes(
	ctx context.Context,
	query string,
) ([]FoursquarePlace, error) {

	if f.APIKey == "" {
		return nil, fmt.Errorf("FOURSQUARE_API_KEY is not set")
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("near", "Nakhon Pathom, Thailand")
	params.Set("categories", "13032")
	params.Set("limit", "20")

	apiURL := "https://places-api.foursquare.com/places/search?" + params.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		apiURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	req.Header.Set("X-Places-Api-Version", "2025-06-17")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("foursquare request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("foursquare returned status %s", resp.Status)
	}

	var result foursquareSearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"failed to decode foursquare response: %w",
			err,
		)
	}

	places := make([]FoursquarePlace, 0, len(result.Results))

	for _, item := range result.Results {
		category := ""

		if len(item.Categories) > 0 {
			category = item.Categories[0].Name
		}

		places = append(places, FoursquarePlace{
			FSQID:     item.FSQPlaceID,
			Name:      item.Name,
			Address:   item.Location.Address,
			Locality:  item.Location.Locality,
			Region:    item.Location.Region,
			Latitude:  item.Latitude,
			Longitude: item.Longitude,
			Category:  category,
			ImageURL:  "",
		})
	}

	return places, nil
}

func NormalizeCafeQuery(query string) string {
	return strings.TrimSpace(query)
}
