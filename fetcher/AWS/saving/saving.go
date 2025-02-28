package saving

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
	"cco-package/fetcher/AWS/models"
)

func ProcessVersionFile(db *gorm.DB, filepath string, regionID uint) error {
	// Open the JSON file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open version file: %v", err)
	}
	defer file.Close()

	// Parse the file content
	var data models.SavingData
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return fmt.Errorf("failed to decode version file: %v", err)
	}

	// Fetch RegionCode from regions table
	var regionCode string
	if err := db.Table("regions").Select("region_code").Where("region_id = ?", regionID).Scan(&regionCode).Error; err != nil || regionCode == "" {
		return fmt.Errorf("failed to fetch region_code for regionID %d: %v", regionID, err)
	}
	log.Printf("Fetched regionCode: %s\n", regionCode)

	// Fetch ProviderID (assuming a single provider for the region)
	var providerID uint
	if err := db.Table("providers").Select("provider_id").Where("provider_name = ?", "AWS").Scan(&providerID).Error; err != nil || providerID == 0 {
		return fmt.Errorf("failed to fetch provider_id for AWS: %v", err)
	}
	log.Printf("Fetched providerID: %d\n", providerID)

	// Process the terms section
	for _, term := range data.TermsPlan.SavingsPlan {
		for _, rate := range term.Rates {
			savingPlan := models.SavingPlan{
				Sku:                    term.Sku,
				DiscountedSku:          rate.DiscountedSku,
				LeaseContractLength:    term.LeaseContractLength.Duration,
				DiscountedRate:         rate.DiscountedRate.Price,
				RegionID:               regionID,
				RegionCode:             regionCode,
				ProviderID:             providerID,
				DiscountedInstanceType: rate.DiscountedInstanceType, // Correct field name
				Unit:                   rate.Unit, // Correctly assigning Unit
			}

			// Insert into the database
			if err := db.Create(&savingPlan).Error; err != nil {
				log.Printf("Failed to insert SavingPlan for SKU %s: %v", term.Sku, err)
			}
		}
	}

	return nil
}
