package runtime

import "time"

// Config holds provider configuration.
type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}
