package services

import (
	"fmt"
	"log"
	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/utils"
	"cco-package/fetcher/Azure/models"
)

func ImportData() error { // fetch and import price data from API
	baseURL := "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview&$filter=serviceName%20eq%20%27Virtual%20Machines%27"
	nextPageLink := baseURL

	// Insert Provider once, since it remains constant
	provider := models.Provider{ProviderName: "Azure"}
	result := config.DB.Where("provider_name = ?", provider.ProviderName).FirstOrCreate(&provider)
	if result.Error != nil {
		return fmt.Errorf("Error inserting provider: %v", result.Error)
	}

	for nextPageLink != "" { // Loop through paginated API responses
		// Fetch data from the current page of the API
		priceData, err := utils.FetchData(nextPageLink)
		if err != nil {
			return fmt.Errorf("error fetching price data: %w", err)
		}

		// Extract pricing items from the JSON response
		items, ok := priceData["Items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid data structure for items")
		}

		// Iterate over each item in the current page
		for _, item := range items {
			data := item.(map[string]interface{})

			// Extract region details
			regionCode, _ := data["location"].(string)
			regionName, _ := data["armRegionName"].(string) // Fetch armRegionName and store in RegionName field

			// Insert Region if it does not exist
			region := models.Region{
				ProviderID: provider.ProviderID,
				RegionCode: regionCode,
				RegionName: regionName, // Maps to arm_region_name in DB
			}
			result = config.DB.Where("region_code = ? AND region_name = ?", region.RegionCode, region.RegionName).FirstOrCreate(&region)
			if result.Error != nil {
				log.Printf("Error inserting region: %v", result.Error)
			}
		}

		// Handle pagination
		if nextLink, ok := priceData["NextPageLink"].(string); ok && nextLink != "" {
			nextPageLink = nextLink
		} else {
			nextPageLink = "" // Exit loop when there are no more pages
		}
	}

	fmt.Println("Data import completed successfully!")
	return nil
}