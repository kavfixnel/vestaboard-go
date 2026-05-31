package vestaboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	cloudBaseURL = "https://cloud.vestaboard.com"
	vbmlBaseURL  = "https://vbml.vestaboard.com"
)

// TransitionType represents the board transition animation.
type TransitionType string

const (
	TransitionClassic TransitionType = "classic"
	TransitionWave    TransitionType = "wave"
	TransitionDrift   TransitionType = "drift"
	TransitionCurtain TransitionType = "curtain"
)

// TransitionSpeed represents the speed of the board transition.
type TransitionSpeed string

const (
	TransitionSpeedGentle TransitionSpeed = "gentle"
	TransitionSpeedFast   TransitionSpeed = "fast"
)

// Characters is a 2D array of character codes representing the board state.
// Flagship: 6 rows × 22 columns. Vestaboard Note: 3 rows × 15 columns.
type Characters [][]int

// Message represents a board message response.
type Message struct {
	ID         string     `json:"_id"`
	Characters Characters `json:"characters"`
	Created    int64      `json:"created"`
}

// Transition represents the board transition settings.
type Transition struct {
	Transition      TransitionType  `json:"transition"`
	TransitionSpeed TransitionSpeed `json:"transitionSpeed"`
}

// Client is a Vestaboard Read/Write API client.
type Client struct {
	token        string
	httpClient   *http.Client
	cloudBaseURL string
	vbmlBaseURL  string
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// withCloudBaseURL overrides the cloud API base URL (for testing).
func withCloudBaseURL(url string) Option {
	return func(c *Client) { c.cloudBaseURL = url }
}

// withVBMLBaseURL overrides the VBML service base URL (for testing).
func withVBMLBaseURL(url string) Option {
	return func(c *Client) { c.vbmlBaseURL = url }
}

// NewClient creates a new Vestaboard client with the given API token.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:        token,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		cloudBaseURL: cloudBaseURL,
		vbmlBaseURL:  vbmlBaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Read returns the current message displayed on the board.
func (c *Client) Read() (*Message, error) {
	req, err := http.NewRequest(http.MethodGet, c.cloudBaseURL+"/", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// WriteText sends a text message to the board.
func (c *Client) WriteText(text string) (*Message, error) {
	return c.writeMessage(map[string]any{"text": text})
}

// WriteTextForced sends a text message to the board, overriding rate limiting.
func (c *Client) WriteTextForced(text string) (*Message, error) {
	return c.writeMessage(map[string]any{"text": text, "forced": true})
}

// WriteCharacters sends a character code array to the board.
func (c *Client) WriteCharacters(chars Characters) (*Message, error) {
	return c.writeMessage(map[string]any{"characters": chars})
}

func (c *Client) writeMessage(body any) (*Message, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.cloudBaseURL+"/", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetTransition returns the current transition settings.
func (c *Client) GetTransition() (*Transition, error) {
	req, err := http.NewRequest(http.MethodGet, c.cloudBaseURL+"/transition", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var t Transition
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// SetTransition updates the board transition settings.
func (c *Client) SetTransition(transition TransitionType, speed TransitionSpeed) (*Transition, error) {
	b, err := json.Marshal(Transition{Transition: transition, TransitionSpeed: speed})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, c.cloudBaseURL+"/transition", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var t Transition
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Format converts a VBML message string into a character code array via the
// VBML service.
func (c *Client) Format(message string) (Characters, error) {
	b, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.vbmlBaseURL+"/format", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var chars Characters
	if err := json.NewDecoder(resp.Body).Decode(&chars); err != nil {
		return nil, err
	}
	return chars, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-Vestaboard-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
}

// APIError represents a non-2xx response from the API.
type APIError struct {
	StatusCode int
	Status     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("vestaboard: API error %s", e.Status)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return &APIError{StatusCode: resp.StatusCode, Status: resp.Status}
}
