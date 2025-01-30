package services

import (
	"data-fetcher/Azure/config"
	"data-fetcher/Azure/models"
	"data-fetcher/Azure/utils"
	"fmt"
	"log"
	"strings"
	"time"
)

func ImportPricesData() error {
	// Prices API URL (Initial URL to start fetching)
	priceApiUrl := "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview&$filter=serviceName%20eq%20%27Virtual%20Machines%27"

	// Loop to handle pagination
	for {
		// Fetch price data
		priceData, err := utils.FetchData(priceApiUrl)
		if err != nil {
			return fmt.Errorf("error fetching price data: %w", err)
		}

		// Parse price items from the response
		priceItems, ok := priceData["Items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid format for price items")
		}

		// Iterate over each price item
		for _, priceItemInterface := range priceItems {
			priceItem := priceItemInterface.(map[string]interface{})

			// Extract required fields from the API response
			skuID, _ := priceItem["skuId"].(string)
			retailPrice, _ := priceItem["retailPrice"].(float64)
			unitOfMeasure, _ := priceItem["unitOfMeasure"].(string)
			effectiveStartDate, _ := priceItem["effectiveStartDate"].(string)

			// Extract only the unit from "1 Hour", "1 Minute", etc.
			unitParts := strings.Fields(unitOfMeasure)
			unit := ""
			if len(unitParts) > 1 {
				unit = unitParts[1] // Fetch "Hour", "Minute", etc.
			}

			// Find the corresponding SKU in the database
			sku := models.SKU{}
			if err := config.DB.Where("id = ?", skuID).First(&sku).Error; err != nil {
    			log.Printf("SKU not found for skuId: %s, skipping...", skuID)
    			continue
			}

			// Parse the effective start date
			effectiveDate, err := time.Parse(time.RFC3339, effectiveStartDate)
			if err != nil {
				log.Printf("Invalid effective start date for skuId: %s, skipping...", skuID)
				continue
			}

			// Create a new Price entry
			price := models.Price{
				SkuID:         sku.ID,  // Directly assign sku.ID (of type uint)
				PricePerUnit:  retailPrice,
				Unit:          unit,
				EffectiveDate: effectiveDate,
			}

			// Insert the Price into the database
			result := config.DB.Create(&price)
			if result.Error != nil {
				log.Printf("Error inserting price for skuId: %s, error: %v", skuID, result.Error)
			} else {
				log.Printf("Price inserted successfully for skuId: %s with price per unit: %.6f", skuID, retailPrice)
			}
		}

		// Check for the next page using the NextPageLink field
		nextPageLink, exists := priceData["NextPageLink"].(string)
		if !exists || nextPageLink == "" {
			log.Println("All pages fetched successfully.")
			break
		}

		// Update the API URL to the next page URL for the next iteration
		priceApiUrl = nextPageLink
	}

	log.Println("Prices data import completed successfully.")
	return nil
}