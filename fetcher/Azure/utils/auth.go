package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)


// logger is a package-level logger that defaults to log.Default().
var logger *log.Logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

// LoadEnv loads the .env file from the Azure folder.
func LoadEnv() {
	envPath := ".env" // Relative path from project root
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Printf("Warning: .env file not found at path: %s\n", envPath)
		return
	}
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Error loading .env file:", err)
	}
}

// GenerateBearerToken generates a bearer token for Azure API access.
func GenerateBearerToken() (string, error) {
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	tenantID := os.Getenv("AZURE_TENANT_ID")

	// Validate environment variables
	if clientID == "" || clientSecret == "" || tenantID == "" {
		missing := []string{}
		if clientID == "" {
			missing = append(missing, "AZURE_CLIENT_ID")
		}
		if clientSecret == "" {
			missing = append(missing, "AZURE_CLIENT_SECRET")
		}
		if tenantID == "" {
			missing = append(missing, "AZURE_TENANT_ID")
		}
		return "", fmt.Errorf("missing required environment variables: %v", missing)
	}

	url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "client_credentials",
		"scope":         "https://management.azure.com/.default",
	}

	// Construct the payload in x-www-form-urlencoded format
	payloadBytes := []byte{}
	for key, value := range payload {
		payloadBytes = append(payloadBytes, []byte(fmt.Sprintf("%s=%s&", key, value))...)
	}
	// Remove the trailing '&'
	payloadBytes = payloadBytes[:len(payloadBytes)-1]

	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("error making token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	token, ok := responseData["access_token"].(string)
	if !ok {
		return "", errors.New("access_token not found in response")
	}

	return token, nil
}