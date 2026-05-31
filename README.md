# vestaboard-go

[![Go Reference](https://pkg.go.dev/badge/github.com/kavfixnel/vestaboard-go.svg)](https://pkg.go.dev/github.com/kavfixnel/vestaboard-go)

A Go client for the [Vestaboard Read/Write API](https://docs.vestaboard.com/docs/read-write-api/introduction).

## Installation

```sh
go get kavfixnel/vestaboard-go
```

## Authentication

Create an API token in the Vestaboard mobile app (Settings) or web app (Developer section). Store it in an environment variable — never hard-code it.

## Usage

```go
client := vestaboard.NewClient(os.Getenv("VESTABOARD_TOKEN"))

// Read the current board state
msg, err := client.Read()

// Send a text message
msg, err = client.WriteText("Hello, World!")

// Send a text message, bypassing the 15-second rate limit
msg, err = client.WriteTextForced("Urgent!")

// Send raw character codes
chars := vestaboard.Characters{
    {0, 1, 2, 3},
    {4, 5, 6, 7},
}
msg, err = client.WriteCharacters(chars)

// Get / set transition animation
transition, err := client.GetTransition()
transition, err  = client.SetTransition(vestaboard.TransitionWave, vestaboard.TransitionSpeedFast)

// Convert a VBML string to character codes via the VBML service
codes, err := client.Format("{63}{63}{63} Hello {63}{63}{63}")
```

## API

### Client

```go
func NewClient(token string, opts ...Option) *Client
```

**Options**

| Option | Description |
|---|---|
| `WithHTTPClient(hc *http.Client)` | Use a custom HTTP client (e.g. custom timeout, proxy) |

### Methods

| Method | Description |
|---|---|
| `Read() (*Message, error)` | Return the current board state |
| `WriteText(text string) (*Message, error)` | Display a text message |
| `WriteTextForced(text string) (*Message, error)` | Display a text message, overriding rate limiting |
| `WriteCharacters(chars Characters) (*Message, error)` | Display raw character codes |
| `GetTransition() (*Transition, error)` | Return current transition settings |
| `SetTransition(t TransitionType, s TransitionSpeed) (*Transition, error)` | Update transition settings |
| `Format(message string) (Characters, error)` | Convert a VBML string to character codes |

### Types

```go
// Characters is a 2D array of character codes.
// Flagship: 6 rows × 22 columns. Vestaboard Note: 3 rows × 15 columns.
type Characters [][]int

type TransitionType  string // TransitionClassic, TransitionWave, TransitionDrift, TransitionCurtain
type TransitionSpeed string // TransitionSpeedGentle, TransitionSpeedFast
```

### Error handling

Non-2xx responses return an `*APIError`:

```go
msg, err := client.WriteText("Hello")
if err != nil {
    var apiErr *vestaboard.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode) // e.g. 429
    }
}
```

## Rate limiting

The Vestaboard cloud API allows one message every 15 seconds. Use `WriteTextForced` to override this for urgent updates.

## License

MIT
