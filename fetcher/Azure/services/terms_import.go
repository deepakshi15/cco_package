package services

import (
	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/models"
	"cco-package/fetcher/Azure/utils"
	"fmt"
	"log"
	"time"
)

func ImportTermsData() error {
	// Prices API base URL
	basePriceApiUrl := "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview&$filter=serviceName%20eq%20%27Virtual%20Machines%27"

	nextPageUrl := basePriceApiUrl
	totalPagesFetched := 0 // Keeps track of how many pages have been fetched from API

	for nextPageUrl != "" { // For pagination
		// Fetch pricing data for the current page
		priceData, err := utils.FetchData(nextPageUrl) // Calls fetch data function to fetch data
		if err != nil {
			return fmt.Errorf("error fetching price data: %w", err)
		}

		// Parse data
		priceItems, ok := priceData["Items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid format for price items")
		}

		// Process each price item
		for _, priceItemInterface := range priceItems {
			priceItem, ok := priceItemInterface.(map[string]interface{})
			if !ok {
				log.Printf("Skipping invalid price item: %v", priceItemInterface)
				continue
			}

			// Extract required fields from the price API
			skuID, _ := priceItem["skuId"].(string)

			// Find the corresponding SKU in the database
			sku := models.SKU{}
			if err := config.DB.Where("sku_code = ?", skuID).First(&sku).Error; err != nil {
				log.Printf("SKU not found for skuId: %s, skipping...", skuID)
				continue
			}

			// Find or create the corresponding price record
			priceRecord := models.Price{}
			priceID := 0 // Initialize as 0, will be updated if price exists
			if err := config.DB.Where("sku_id = ?", sku.ID).First(&priceRecord).Error; err != nil {
				// Insert the price record if it doesn't exist
				priceRecord = models.Price{
					SkuID: sku.ID, // Ensure this matches your foreign key type
				}
				if err := config.DB.Create(&priceRecord).Error; err != nil {
					log.Printf("Error creating price record for skuId: %s, error: %v", skuID, err)
					continue
				}
				priceID = priceRecord.PriceID
				log.Printf("Created new price record for skuId: %s", skuID)
			} else {
				priceID = priceRecord.PriceID
			}

			// Extract savingsPlan from the price API
			savingsPlans, ok := priceItem["savingsPlan"].([]interface{})
			if !ok {
				log.Printf("No savings plan available for skuId: %s, skipping...", skuID)
				continue
			}

			// Process each savings plan
			for _, planInterface := range savingsPlans {
				plan, ok := planInterface.(map[string]interface{})
				if !ok {
					log.Printf("Skipping invalid savings plan for skuId: %s", skuID)
					continue
				}

				leaseContractLength, _ := plan["term"].(string)

				// Create a new Term entry
				term := models.Term{
					PriceID:             uint(priceID),                // Convert int to uint (corrected)
					SkuID:               sku.ID,                       // SkuID is an int in your struct (already correctly set)
					PurchaseOption:      nil,                           // Set to NULL as there is no data for it
					OfferingClass:       nil,                           // Set to NULL as there is no data for it
					LeaseContractLength: &leaseContractLength,         // Nullable field (correct)
					CreatedDate:         time.Now(),                    // Automatically generated
					ModifiedDate:        time.Now(),                    // Automatically generated
					DisableFlag:         false,                         // Default value
				}

				// Insert the Term into the database
				result := config.DB.Create(&term)
				if result.Error != nil {
					log.Printf("Error inserting term for skuId: %s, error: %v", skuID, result.Error)
				} else {
				continue
				}
			}
		}

		// Increment total pages fetched
		totalPagesFetched++

		// Get the next page URL
		nextPageUrl, _ = priceData["NextPageLink"].(string)
		log.Printf("Next page URL: %s", nextPageUrl)
	}

	log.Println("Terms data import completed successfully.")
	return nil
}
