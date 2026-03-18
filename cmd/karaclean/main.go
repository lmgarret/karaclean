package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/karakeep"
)

func main() {
	// Step 1: Load config
	path := config.ResolvePath("")
	_, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Read required env vars
	karakeepURL, err := requireEnv("KARAKEEP_URL")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	apiKey, err := requireEnv("KARAKEEP_API_KEY")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Construct API client
	client, err := karakeep.NewKarakeepClient(karakeepURL, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Validate auth token on startup (CONF-03)
	if err := client.CheckAuth(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("authenticated successfully")
}

// requireEnv returns the value of the named environment variable,
// or an error if it is unset or empty.
func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return val, nil
}
