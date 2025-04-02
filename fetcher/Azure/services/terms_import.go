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
	totalPagesFetched := 0 // Tracks pages fetched

	for nextPageUrl != "" { // Pagination loop
		// Fetch pricing data for the current page
		priceData, err := utils.FetchData(nextPageUrl)
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

			// Extract required fields from the API
			skuID, _ := priceItem["skuId"].(string)

			// Find the corresponding SKU in the database using `sku_code`
			sku := models.SKU{}
			if err := config.DB.Where("sku_code = ?", skuID).First(&sku).Error; err != nil {
				log.Printf("SKU not found for sku_code: %s, skipping...", skuID)
				continue
			}

			// Find or create the corresponding price record
			priceRecord := models.Price{}
			var priceID uint // Default is 0, will be updated if price exists
			if err := config.DB.Where("sku_id = ?", sku.ID).First(&priceRecord).Error; err != nil {
				// Insert the price record if it doesn't exist
				priceRecord = models.Price{
					SkuID: sku.ID, // Correctly assign the uint ID
				}
				if err := config.DB.Create(&priceRecord).Error; err != nil {
					log.Printf("Error creating price record for sku_code: %s, error: %v", skuID, err)
					continue
				}
				priceID = priceRecord.PriceID
				log.Printf("Created new price record for sku_code: %s", skuID)
			} else {
				priceID = priceRecord.PriceID
			}

			// Extract savingsPlan from the price API
			savingsPlans, ok := priceItem["savingsPlan"].([]interface{})
			if !ok {
				log.Printf("No savings plan available for sku_code: %s, skipping...", skuID)
				continue
			}

			// Process each savings plan
			for _, planInterface := range savingsPlans {
				plan, ok := planInterface.(map[string]interface{})
				if !ok {
					log.Printf("Skipping invalid savings plan for sku_code: %s", skuID)
					continue
				}

				leaseContractLength, _ := plan["term"].(string)

				// Ensure nullable fields are properly assigned
				var leaseContractLengthPtr *string = nil
				if leaseContractLength != "" {
					leaseContractLengthPtr = &leaseContractLength
				}

				// Create a new Term entry
				term := models.Term{
					PriceID:             priceID,
					SkuID:               sku.ID,
					PurchaseOption:      nil,                     // Set to NULL
					OfferingClass:       nil,                     // Set to NULL
					LeaseContractLength: leaseContractLengthPtr,  // Nullable field
					CreatedDate:         time.Now(),
					ModifiedDate:        time.Now(),
					DisableFlag:         false,
				}

				// Insert the Term into the database
				result := config.DB.Create(&term)
				if result.Error != nil {
					log.Printf("Error inserting term for sku_code: %s, error: %v", skuID, result.Error)
				} else {
					log.Printf("Term inserted successfully for sku_code: %s with lease_contract_length: %s", skuID, leaseContractLength)
				}
			}
		}

		// Increment total pages fetched
		totalPagesFetched++

		// Get the next page URL
		nextPageUrl, _ = priceData["NextPageLink"].(string)
		log.Printf("Next page URL: %s", nextPageUrl)

		// Optional delay to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Println("Terms data import completed successfully.")
	return nil
}
