package config

import (
	"fmt"
	"os/exec"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var AuthToken string

func ConnectDatabase() {
	// Set up DB connection
	dsn := "host=localhost user=postgres password=password dbname=temp_db port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	DB = database

	// Fetch access token using gcloud
	token, err := getGcloudAccessToken()
	if err != nil {
		panic(fmt.Sprintf("Failed to get GCP access token: %v", err))
	}
	AuthToken = "Bearer " + token
}

func getGcloudAccessToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}