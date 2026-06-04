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

	"github.com/lmgarret/karaclean/internal/config"
	"github.com/lmgarret/karaclean/internal/engine"
	"github.com/lmgarret/karaclean/internal/karakeep"
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
	cfg := loadConfig(configPath)

	// Step 1.5: Resolve dry-run mode (flag > env > config)
	dryRun := resolveDryRunFromFlags(dryRunFlag, cfg.DryRun)
	if dryRun {
		log.Println("dry-run mode enabled -- no mutations will be executed")
	}

	// Steps 2-4.25: Construct and validate API client
	client := initClient(cfg)

	// Step 4.5: Create notifier for notification dispatch
	var notifier engine.Notifier
	if cfg.Notifications != nil {
		notifier = &engine.ShoutrrrNotifier{}
	}

	// Step 5: Resolve timezone
	loc := resolveTimezone(cfg.Timezone)

	// Step 6: Create signal-aware context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Step 7: Run immediately at startup (synchronous, before cron starts)
	summary, err := engine.Run(ctx, client, cfg.Rules, dryRun, cfg.Notifications, notifier)
	if err != nil {
		exitf("error: %v", err)
	}
	log.Printf("run complete: %s", summary)

	// Step 8: Set up cron scheduler and block until signal
	runScheduler(ctx, cfg, client, dryRun, notifier, loc)
}

// loadConfig loads and validates the configuration file.
// Passes a ShoutrrrNotifier to Load so that config errors can be sent
// via the configured notification channel when notifyOnError is true.
func loadConfig(configPath string) *config.Config {
	path := config.ResolvePath(configPath)
	cfg, err := config.Load(path, &engine.ShoutrrrNotifier{})
	if err != nil {
		exitf("error: %v", err)
	}
	return cfg
}

// resolveDryRunFromFlags determines dry-run mode from CLI flags, env, and config.
func resolveDryRunFromFlags(dryRunFlag bool, configVal bool) bool {
	dryRunFlagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "dry-run" {
			dryRunFlagSet = true
		}
	})
	return resolveDryRun(dryRunFlagSet, dryRunFlag, os.Getenv("KARACLEAN_DRY_RUN"), configVal)
}

// initClient creates, authenticates, and validates the Karakeep API client.
func initClient(cfg *config.Config) engine.KarakeepAPI {
	karakeepURL, err := requireEnv("KARAKEEP_URL")
	if err != nil {
		exitf("error: %v", err)
	}
	apiKey, err := requireEnv("KARAKEEP_API_KEY")
	if err != nil {
		exitf("error: %v", err)
	}

	client, err := karakeep.NewKarakeepClient(karakeepURL, apiKey)
	if err != nil {
		exitf("error: %v", err)
	}

	if err := client.CheckAuth(context.Background()); err != nil {
		exitf("error: %v", err)
	}

	if listNames := cfg.CollectListNames(); len(listNames) > 0 {
		if err := validateListNames(context.Background(), client, listNames); err != nil {
			exitf("error: %v", err)
		}
	}

	return client
}

// resolveTimezone returns the timezone location from config, defaulting to UTC.
func resolveTimezone(tz string) *time.Location {
	if tz != "" {
		loc, _ := time.LoadLocation(tz) // already validated by config.Validate()
		return loc
	}
	log.Println("WARNING: timezone not set, defaulting to UTC")
	return time.UTC
}

// runScheduler sets up the cron scheduler and blocks until context cancellation.
func runScheduler(ctx context.Context, cfg *config.Config, client engine.KarakeepAPI, dryRun bool, notifier engine.Notifier, loc *time.Location) {
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

	<-ctx.Done()
	log.Println("received signal, shutting down")

	stopCtx := c.Stop()
	<-stopCtx.Done()
}

// exitf prints a formatted error message to stderr and exits with code 1.
func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
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

// validateListNames checks that all configured list names exist in Karakeep.
// Returns an error listing all missing names (D-13).
func validateListNames(ctx context.Context, client engine.KarakeepAPI, listNames []string) error {
	lists, err := client.ListLists(ctx)
	if err != nil {
		return fmt.Errorf("validating list names: %w", err)
	}
	existing := make(map[string]bool, len(lists))
	for _, l := range lists {
		existing[l.Name] = true
	}
	var missing []string
	for _, name := range listNames {
		if !existing[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("list names not found in Karakeep: %v", missing)
	}
	return nil
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
