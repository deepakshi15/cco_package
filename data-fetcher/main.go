package main

import (
    "data-fetcher/AWS"
    "data-fetcher/Azure"
    "data-fetcher/config"
    "sync"
)

func main() {
    // Initialize database
    config.ConnectDatabase()

    // Use WaitGroup to run AWS and Azure concurrently
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        AWS.RunAWS()
    }()

    go func() {
        defer wg.Done()
        Azure.RunAzure()
    }()

    wg.Wait()
    println("AWS and Azure tasks completed!")
}
