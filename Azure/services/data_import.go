package services

import (
	"fmt"
	"log"
	"cco_backend/config"
	"cco_backend/models"
	"cco_backend/utils"
)

func ImportData() error { // fetch and import price data from API
	baseURL := "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview&$filter=serviceName%20eq%20%27Virtual%20Machines%27"
	nextPageLink := baseURL

	// Insert Provider and Service once, as they remain constant throughout the data
	provider := models.Provider{ProviderName: "Azure"}
	result := config.DB.Where("provider_name = ?", provider.ProviderName).FirstOrCreate(&provider)
	if result.Error != nil {
		return fmt.Errorf("Error inserting provider: %v", result.Error)
	} else {
		log.Printf("Provider inserted or already exists: %v", provider.ProviderName)
	}

	for nextPageLink != "" { // loops through the API's paginated responses
		// Fetch data from the current page of the price API
		priceData, err := utils.FetchData(nextPageLink)
		if err != nil {
			return fmt.Errorf("error fetching price data: %w", err)
		}

		// Extract items from the JSON response - contains array of pricing data
		items, ok := priceData["Items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid data structure for items")
		}

		// Iterate over each item in the current page
		for _, item := range items {
			data := item.(map[string]interface{})

			// For region table
			regionCode, _ := data["location"].(string)

			// Insert Region if not exists
			region := models.Region{
				ProviderID: provider.ProviderID,
				RegionCode: regionCode,
			}
			result = config.DB.Where("region_code = ?", region.RegionCode).FirstOrCreate(&region)
			if result.Error != nil {
				log.Printf("Error inserting region: %v", result.Error)
			} else {
				log.Printf("Region inserted or already exists: %v", region.RegionCode)
			}
		}

		// Update the nextPageLink for the next iteration (handles pagination)
		if nextLink, ok := priceData["NextPageLink"].(string); ok && nextLink != "" {
			nextPageLink = nextLink
		} else {
			nextPageLink = "" // Exit the loop if there's no next page
		}
	}

	fmt.Println("Data import completed successfully!")
	return nil
}
