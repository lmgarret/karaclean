package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/karakeep"
)

func main() {
	// Step 0: Parse CLI flags
	var configPath string
	var dryRunFlag bool
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "enable dry-run mode (no mutations)")
	flag.Parse()

	// Step 1: Load config
	path := config.ResolvePath(configPath)
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Step 1.5: Resolve dry-run mode (flag > env > config)
	dryRunFlagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "dry-run" {
			dryRunFlagSet = true
		}
	})
	dryRun := resolveDryRun(dryRunFlagSet, dryRunFlag, os.Getenv("KARACLEAN_DRY_RUN"), cfg.DryRun)
	if dryRun {
		log.Println("dry-run mode enabled -- no mutations will be executed")
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

// resolveDryRun determines dry-run mode using precedence: flag > env var > config field.
// flagSet indicates whether --dry-run was explicitly passed on the command line.
func resolveDryRun(flagSet bool, flagVal bool, envVal string, configVal bool) bool {
	if flagSet {
		return flagVal
	}
	if envVal != "" {
		return envVal == "true" || envVal == "1"
	}
	return configVal
}
