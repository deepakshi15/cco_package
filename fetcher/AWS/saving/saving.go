package saving

import (
	"encoding/json"
	"log"
	"os"
	"gorm.io/gorm"
	"cco-package/fetcher/AWS/models"
)

func ProcessVersionFile(db *gorm.DB, filepath string, regionID uint) error {
	// Open the JSON file
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to open version file: %v", err)
	}
	defer file.Close()

	// Parse the file content
	var data models.SavingData
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Fatalf("Failed to decode version file: %v", err)
	}

	// Process the terms section
	for _, term := range data.TermsPlan.SavingsPlan {
		for _, rate := range term.Rates {
			savingPlan := models.SavingPlan{
				Sku:                 term.Sku,
				DiscountedSku:       rate.DiscountedSku,
				LeaseContractLength: term.LeaseContractLength.Duration,
				DiscountedRate:      rate.DiscountedRate.Price,
				RegionID:            regionID,
			}

			// Insert into the database
			if err := db.Create(&savingPlan).Error; err != nil {
				log.Printf("Failed to insert SavingPlan for SKU %s: %v", term.Sku, err)
			} else {
				continue
			}
		}
	}

	return nil
}