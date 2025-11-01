package config

import "sync"

// Config holds the application's configuration.
type Config struct {
	ProxyPort     int
	WebSocketPort int
	Pause         bool
	FilterMozilla bool
	mu            sync.Mutex
}

var (
	instance *Config
	once     sync.Once
)

// SetPause sets the pause value in a thread-safe way.
func (c *Config) SetPause(pause bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Pause = pause
}

// GetInstance returns the singleton instance of the Config.
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{
			// Default values
			ProxyPort:     8181,
			WebSocketPort: 8182, // Default WebSocket port
			Pause:         false,
			FilterMozilla: true,
		}
	})
	return instance
}
