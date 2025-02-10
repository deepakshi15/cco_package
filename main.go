package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
	"cco-package/fetcher"
	"cco-package/updatedatabase"
	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/utils" // Import Azure utils package for auth functions.
)

// Retry settings
const (
	maxRetries   = 5               // Maximum number of retry attempts
	initialDelay = 2 * time.Second // Initial retry delay
)

// executeWithRetry attempts to execute a function with retries on failure.
func executeWithRetry(task func() error, taskName string) {
	var delay = initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := task()
		if err == nil {
			log.Printf("%s completed successfully.\n", taskName)
			return
		}

		log.Printf("Attempt %d failed for %s: %v. Retrying in %v...\n", attempt, taskName, err, delay)

		// Apply exponential backoff with jitter
		time.Sleep(delay + time.Duration(rand.Intn(1000))*time.Millisecond)

		// Double the delay for the next attempt (up to a reasonable limit)
		if delay < 30*time.Second {
			delay *= 2
		}
	}

	log.Printf("All retries failed for %s. Moving on...\n", taskName)
}

// Wrappers to convert functions to return an error.
func dataFetcher() error {
	err := fetcher.Fetcher()
	if err != nil {
		return err
	}
	return nil
}

func updateDatabaseTask() error {
	err := updatedatabase.Updatedatabase() // Ensure updatedatabase.Updatedatabase() returns an error.
	if err != nil {
		return err
	}
	return nil
}

func runTask() {
	log.Println("Task started at:", time.Now())
	// Run AWS fetch with retry.
	executeWithRetry(dataFetcher, "Fetching Data")
	// Run Database update with retry.
	executeWithRetry(updateDatabaseTask, "Update Database")
	log.Println("Task completed at:", time.Now())
}

func main() {
	// Initialize the logger and clear the logfile before appending new logs.
	logger, err := config.InitializeLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Redirect the global logger to write to the same file.
	log.SetOutput(logger.Writer())

	// Pass the logger to other packages.
	fetcher.SetLogger(logger)
	utils.SetLogger(logger)

	// Load the .env file and ensure logs are properly captured.
	utils.LoadEnv()

	// Set up the cron job.
	c := cron.New()
	_, err = c.AddFunc("@every 1m", runTask)
	if err != nil {
		logger.Fatal("Error scheduling cron job:", err)
	}
	logger.Println("Cron job started, will run every minute.")
	c.Start()

	// Block the main goroutine indefinitely so the cron scheduler keeps running.
	select {}
}