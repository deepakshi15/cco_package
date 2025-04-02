package fetcher

import (
	//"cco-package/fetcher/AWS"
	"cco-package/fetcher/Azure"
	"cco-package/fetcher/config"
	"log"
	"sync"
)

// logger is a package-level variable that can be set from main.
var logger *log.Logger

// SetLogger allows you to inject a logger into the fetcher package.
func SetLogger(l *log.Logger) {
	logger = l
}

// DoSomething logs a sample message using the provided logger.
func DoSomething(l *log.Logger) {
	l.Println("This is a log message from the fetcher package.")
	// Your additional code here...
}

// Fetcher runs the AWS and Azure tasks concurrently and returns the first error (if any).
func Fetcher() error {
	if logger != nil {
		logger.Println("Starting Fetcher")
	}

	// Initialize database via the config package.
	if err := config.ConnectDatabase(); err != nil {
		return err
	}

	// Use a WaitGroup to run AWS and Azure concurrently.
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2) // Buffered channel to collect errors

	// go func() {
	// 	defer wg.Done()
	// 	if err := AWS.RunAWS(); err != nil {
	// 		errChan <- err
	// 	}
	// }()

	go func() {
		defer wg.Done()
		if err := Azure.RunAzure(); err != nil {
			errChan <- err
		}
	}()

	// Wait for both goroutines to complete.
	wg.Wait()
	close(errChan)

	// Collect and return the first error encountered (if any).
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Println("AWS and Azure tasks completed successfully!")
	} else {
		println("AWS and Azure tasks completed successfully!")
	}

	return nil
}