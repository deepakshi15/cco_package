package GCP
import (
	"fmt"

	"cco-package/fetcher/GCP/config"
	"cco-package/fetcher/GCP/services"
)

func RunGCP() error {
	// Connect to DB (you might want to add error handling if ConnectDatabase ever returns an error)
	config.ConnectDatabase()

	// Step 1: Fetch and store regions
	if err := services.FetchAndStoreRegions(); err != nil {
		return fmt.Errorf("error syncing regions: %w", err)
	}

	// Step 2: Fetch and store SKUs
	if err := services.FetchAndInsertSkus(); err != nil {
		return fmt.Errorf("error syncing SKUs: %w", err)
	}

	return nil
}
