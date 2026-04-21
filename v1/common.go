package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/robomotionio/robomotion-go/runtime"
)

const (
	baseURL = "https://roundups.ai/api/v1"
)

// parseStringArray converts various array-like Go values into []string.
// Handles []interface{} (with string elements), []string, and single strings.
// Empty strings are filtered out. Non-string elements in arrays return an error.
func parseStringArray(v interface{}) ([]string, error) {
	switch arr := v.(type) {
	case []string:
		result := make([]string, 0, len(arr))
		for _, s := range arr {
			if s != "" {
				result = append(result, s)
			}
		}
		return result, nil
	case []interface{}:
		result := make([]string, 0, len(arr))
		for i, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("array element at index %d is not a string", i)
			}
			if s != "" {
				result = append(result, s)
			}
		}
		return result, nil
	case string:
		if arr != "" {
			return []string{arr}, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("expected string array, got %T", v)
	}
}

// RoundupsClient handles HTTP communication with the Roundups API.
type RoundupsClient struct {
	httpClient *http.Client
	apiKey     string
}

// NewRoundupsClient creates a new API client.
func NewRoundupsClient(apiKey string) *RoundupsClient {
	return &RoundupsClient{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		apiKey: apiKey,
	}
}

// CreateRoundupRequest represents the POST /api/v1/roundups request body.
type CreateRoundupRequest struct {
	Headline            string        `json:"headline,omitempty"`
	TargetAudience      string        `json:"target_audience,omitempty"`
	Keywords            string        `json:"keywords,omitempty"`
	ProductType         string        `json:"product_type,omitempty"`
	ProductsCount       int           `json:"products_count,omitempty"`
	ProductsSearchQuery []string      `json:"products_search_queries,omitempty"`
	AmazonProductASINs  []string      `json:"amazon_product_asins,omitempty"`
	ProductURLs         []string      `json:"product_urls,omitempty"`
	Styles              *StyleOptions `json:"styles,omitempty"`
}

// StyleOptions represents content settings for roundup generation.
type StyleOptions struct {
	ToneOfVoice        string `json:"tone_of_voice,omitempty"`
	Language           string `json:"language,omitempty"`
	ComparisonTable    *bool  `json:"comparison_table_enabled,omitempty"`
	PointOfView        string `json:"point_of_view,omitempty"`
	CustomCTA          string `json:"custom_cta,omitempty"`
	IncludePricing     *bool  `json:"include_pricing,omitempty"`
	IncludeRating      *bool  `json:"include_rating,omitempty"`
	OptimizeOutputFor  string `json:"optimize_output_for,omitempty"`
	LLMModel           string `json:"llm_model,omitempty"`
	CoverImageStyle    string `json:"cover_image_style,omitempty"`
	ProductCount       int    `json:"product_count,omitempty"`
	VisualStyle        string `json:"visual_style,omitempty"`
	LayoutStyle        string `json:"layout_style,omitempty"`
	TemplateType       string `json:"template_type,omitempty"`
}

// CreateRoundupResponse represents the response from creating a roundup.
type CreateRoundupResponse struct {
	ID        int    `json:"id"`
	Headline  string `json:"headline"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Article   *Article `json:"article,omitempty"`
	Errors    *string  `json:"errors,omitempty"`
}

// Article represents the generated roundup content.
type Article struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	FeaturedImage   string `json:"featured_image"`
	MetaDescription string `json:"meta_description"`
	CreatedAt       string `json:"created_at"`
}

// FetchRoundupResponse represents the response from fetching a roundup.
type FetchRoundupResponse struct {
	ID        int       `json:"id"`
	Headline  string    `json:"headline"`
	State     string    `json:"state"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Article   *Article  `json:"article,omitempty"`
	Errors    *string   `json:"errors,omitempty"`
}

// CreateRoundup sends a POST request to create a new roundup.
func (c *RoundupsClient) CreateRoundup(req *CreateRoundupRequest) (*CreateRoundupResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", baseURL+"/roundups", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, runtime.NewError("ErrAPIError", fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
	}

	var result CreateRoundupResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// FetchRoundup sends a GET request to fetch a roundup by ID.
func (c *RoundupsClient) FetchRoundup(id int) (*FetchRoundupResponse, error) {
	url := fmt.Sprintf("%s/roundups/%d", baseURL, id)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, runtime.NewError("ErrAPIError", fmt.Sprintf("API error %d: %s", resp.StatusCode, string(respBody)))
	}

	var result FetchRoundupResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}


