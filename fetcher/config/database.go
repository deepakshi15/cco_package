package config

import (
    "fmt"
    "log"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// API Links
const (
    BaseURL         = "https://pricing.us-east-1.amazonaws.com"
    RegionURL       = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/region_index.json"
    SavingRegionURL = "https://pricing.us-east-1.amazonaws.com/savingsPlan/v1.0/aws/AWSComputeSavingsPlan/current/region_index.json"
    DbConnStr       = "host=localhost user=postgres password=password dbname=temp_db sslmode=disable"
    PriceListPath   = "./price-list"
)

var DB *gorm.DB

// ConnectDatabase establishes a connection to the PostgreSQL database.
func ConnectDatabase() error {
    var err error
    DB, err = gorm.Open(postgres.Open(DbConnStr), &gorm.Config{})
    if err != nil {
        log.Printf("Error connecting to the database: %v", err)
        return err // Return the error instead of terminating the program
    }
    fmt.Println("Database connected successfully!")
    return nil // Return nil if the connection is successful
}

// InitializeLogger initializes the logger and clears the logfile on every run.
func InitializeLogger() (*log.Logger, error) {
	// Use os.Create to truncate the file or create a new one if it doesn't exist.
	file, err := os.Create("logfile.log")
	if err != nil {
		return nil, err
	}

	// Create a new logger instance that writes to logfile.log.
	logger := log.New(file, "PREFIX: ", log.Ldate|log.Ltime|log.Lshortfile)
	return logger, nil
}