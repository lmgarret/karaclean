package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/engine"
	"github.com/lm/karaclean/internal/karakeep"
	"github.com/robfig/cron/v3"
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

	// Step 4.5: Create notifier for notification dispatch
	var notifier engine.Notifier
	if cfg.Notifications != nil {
		notifier = &engine.ShoutrrrNotifier{}
	}

	// Step 5: Resolve timezone
	loc := time.UTC
	if cfg.Timezone != "" {
		loc, _ = time.LoadLocation(cfg.Timezone) // already validated by config.Validate()
	} else {
		log.Println("WARNING: timezone not set, defaulting to UTC")
	}

	// Step 6: Create signal-aware context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Step 7: Run immediately at startup (synchronous, before cron starts)
	summary, err := engine.Run(ctx, client, cfg.Rules, dryRun, cfg.Notifications, notifier)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("run complete: %s", summary)

	// Step 8: Set up cron scheduler
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	c := cron.New(
		cron.WithParser(parser),
		cron.WithLocation(loc),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
	)
	if _, err := c.AddFunc(cfg.Schedule, func() {
		summary, err := engine.Run(ctx, client, cfg.Rules, dryRun, cfg.Notifications, notifier)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		log.Printf("run complete: %s", summary)
	}); err != nil {
		log.Fatalf("failed to register cron job: %v", err)
	}

	c.Start()
	entries := c.Entries()
	if len(entries) > 0 {
		log.Printf("next run at %s", entries[0].Next.Format(time.RFC3339))
	}

	// Step 9: Block until signal
	<-ctx.Done()
	log.Println("received signal, shutting down")

	// Step 10: Graceful stop -- wait for in-progress job to finish
	stopCtx := c.Stop()
	<-stopCtx.Done()
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
