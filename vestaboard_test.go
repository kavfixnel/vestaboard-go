package vestaboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testToken = "test-token-abc"

// newTestClient wires a Client to talk to the provided httptest servers.
// cloudSrv handles cloud API calls; vbmlSrv handles VBML format calls.
// Pass nil for a server you don't need in a given test.
func newTestClient(cloudSrv, vbmlSrv *httptest.Server) *Client {
	opts := []Option{}
	if cloudSrv != nil {
		opts = append(opts, withCloudBaseURL(cloudSrv.URL))
	}
	if vbmlSrv != nil {
		opts = append(opts, withVBMLBaseURL(vbmlSrv.URL))
	}
	return NewClient(testToken, opts...)
}

// assertToken checks that the request carries the expected API token.
func assertToken(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("X-Vestaboard-Token"); got != testToken {
		t.Errorf("X-Vestaboard-Token = %q, want %q", got, testToken)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// --- Read ---

func TestRead(t *testing.T) {
	want := Message{
		ID:      "msg-1",
		Created: 1700000000,
		Characters: Characters{
			{0, 1, 2},
			{3, 4, 5},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertToken(t, r)
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		writeJSON(w, want)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if len(got.Characters) != len(want.Characters) {
		t.Errorf("Characters rows = %d, want %d", len(got.Characters), len(want.Characters))
	}
}

func TestRead_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	_, err := c.Read()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

// --- WriteText ---

func TestWriteText(t *testing.T) {
	want := Message{ID: "msg-2", Created: 1700000001}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertToken(t, r)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["text"] != "Hello World" {
			t.Errorf("text = %v, want %q", body["text"], "Hello World")
		}
		if _, hasForced := body["forced"]; hasForced {
			t.Error("forced field should not be present for WriteText")
		}
		writeJSON(w, want)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.WriteText("Hello World")
	if err != nil {
		t.Fatalf("WriteText() error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
}

// --- WriteTextForced ---

func TestWriteTextForced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["forced"] != true {
			t.Errorf("forced = %v, want true", body["forced"])
		}
		writeJSON(w, Message{ID: "msg-forced"})
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.WriteTextForced("Urgent!")
	if err != nil {
		t.Fatalf("WriteTextForced() error: %v", err)
	}
	if got.ID != "msg-forced" {
		t.Errorf("ID = %q, want %q", got.ID, "msg-forced")
	}
}

// --- WriteCharacters ---

func TestWriteCharacters(t *testing.T) {
	chars := Characters{{1, 2, 3}, {4, 5, 6}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertToken(t, r)
		var body struct {
			Characters Characters `json:"characters"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(body.Characters) != len(chars) {
			t.Errorf("characters rows = %d, want %d", len(body.Characters), len(chars))
		}
		writeJSON(w, Message{ID: "msg-chars"})
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.WriteCharacters(chars)
	if err != nil {
		t.Fatalf("WriteCharacters() error: %v", err)
	}
	if got.ID != "msg-chars" {
		t.Errorf("ID = %q, want %q", got.ID, "msg-chars")
	}
}

// --- GetTransition ---

func TestGetTransition(t *testing.T) {
	want := Transition{Transition: TransitionWave, TransitionSpeed: TransitionSpeedFast}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertToken(t, r)
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/transition" {
			t.Errorf("path = %q, want /transition", r.URL.Path)
		}
		writeJSON(w, want)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.GetTransition()
	if err != nil {
		t.Fatalf("GetTransition() error: %v", err)
	}
	if got.Transition != want.Transition {
		t.Errorf("Transition = %q, want %q", got.Transition, want.Transition)
	}
	if got.TransitionSpeed != want.TransitionSpeed {
		t.Errorf("TransitionSpeed = %q, want %q", got.TransitionSpeed, want.TransitionSpeed)
	}
}

// --- SetTransition ---

func TestSetTransition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertToken(t, r)
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		var body Transition
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Transition != TransitionDrift {
			t.Errorf("Transition = %q, want %q", body.Transition, TransitionDrift)
		}
		if body.TransitionSpeed != TransitionSpeedGentle {
			t.Errorf("TransitionSpeed = %q, want %q", body.TransitionSpeed, TransitionSpeedGentle)
		}
		writeJSON(w, body)
	}))
	defer srv.Close()

	c := newTestClient(srv, nil)
	got, err := c.SetTransition(TransitionDrift, TransitionSpeedGentle)
	if err != nil {
		t.Fatalf("SetTransition() error: %v", err)
	}
	if got.Transition != TransitionDrift {
		t.Errorf("Transition = %q, want %q", got.Transition, TransitionDrift)
	}
}

// --- Format ---

func TestFormat(t *testing.T) {
	wantChars := Characters{{65, 66, 67}, {68, 69, 70}}

	vbmlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/format" {
			t.Errorf("path = %q, want /format", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["message"] != "ABC" {
			t.Errorf("message = %q, want %q", body["message"], "ABC")
		}
		writeJSON(w, wantChars)
	}))
	defer vbmlSrv.Close()

	c := newTestClient(nil, vbmlSrv)
	got, err := c.Format("ABC")
	if err != nil {
		t.Fatalf("Format() error: %v", err)
	}
	if len(got) != len(wantChars) {
		t.Errorf("rows = %d, want %d", len(got), len(wantChars))
	}
}

func TestFormat_APIError(t *testing.T) {
	vbmlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer vbmlSrv.Close()

	c := newTestClient(nil, vbmlSrv)
	_, err := c.Format("bad input")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
}

// --- APIError.Error() ---

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 429, Status: "429 Too Many Requests"}
	want := "vestaboard: API error 429 Too Many Requests"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// --- WithHTTPClient ---

func TestWithHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, Message{ID: "msg-custom-transport"})
	}))
	defer srv.Close()

	custom := &http.Client{Timeout: srv.Client().Timeout}
	c := NewClient(testToken, withCloudBaseURL(srv.URL), WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not set the custom HTTP client")
	}
	got, err := c.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.ID != "msg-custom-transport" {
		t.Errorf("ID = %q, want %q", got.ID, "msg-custom-transport")
	}
}

// --- Transport error paths ---
// Closing the server before the call forces a connection-refused error,
// exercising the httpClient.Do failure branch in each method.

func TestRead_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := newTestClient(srv, nil)
	srv.Close()
	if _, err := c.Read(); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}

func TestWriteMessage_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := newTestClient(srv, nil)
	srv.Close()
	if _, err := c.WriteText("hi"); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}

func TestGetTransition_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := newTestClient(srv, nil)
	srv.Close()
	if _, err := c.GetTransition(); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}

func TestSetTransition_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := newTestClient(srv, nil)
	srv.Close()
	if _, err := c.SetTransition(TransitionClassic, TransitionSpeedGentle); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}

func TestFormat_TransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	c := newTestClient(nil, srv)
	srv.Close()
	if _, err := c.Format("hi"); err == nil {
		t.Fatal("expected transport error, got nil")
	}
}
