package main

import (
	"os"
)

// GetTMDBAPIKey retrieves the TMDB API key from environment variable or returns placeholder
func GetTMDBAPIKey() string {
	key := os.Getenv("TMDB_API_KEY")
	if key == "" {
		// You can also read from a config file or use build tags
		key = "YOUR_API_KEY_HERE" // Fallback placeholder
	}
	return key
}