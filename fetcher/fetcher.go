package fetcher

import (
    "cco-package/fetcher/AWS"
    "cco-package/fetcher/Azure"
    "cco-package/fetcher/config"
    "sync"
)

func Fetcher() error {
    // Initialize database
    if err := config.ConnectDatabase(); err != nil {
        return err
    }

    // Use WaitGroup to run AWS and Azure concurrently
    var wg sync.WaitGroup
    wg.Add(2)

    errChan := make(chan error, 2) // Buffered channel to collect errors

    go func() {
        defer wg.Done()
        if err := AWS.RunAWS(); err != nil {
            errChan <- err
        }
    }()

    go func() {
        defer wg.Done()
        if err := Azure.RunAzure(); err != nil {
            errChan <- err
        }
    }()

    // Wait for both goroutines to complete
    wg.Wait()
    close(errChan)

    // Collect the first error if any
    for err := range errChan {
        if err != nil {
            return err
        }
    }

    println("AWS and Azure tasks completed successfully!")
    return nil
}
