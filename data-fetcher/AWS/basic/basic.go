package basic

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"gorm.io/gorm"
	"data-fetcher/AWS/models"
    "data-fetcher/AWS/utils"
)
func ProcessCurrentVersionFile(db *gorm.DB, filepath string, regionID uint) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to open current version file: %v", err)
	}
	defer file.Close()

	var data models.PricingData

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Fatalf("Failed to decode current version file: %v", err)
	}

	// Convert map to slice and call function to process products (SKUs)
	productsSlice := mapToSlice(data.Products)
	processProducts(db, productsSlice, regionID)

	// Call function to process terms

	processTerms(db, data.Terms["OnDemand"])
	processTerms(db, data.Terms["Reserved"])
}

// Helper function to convert map to slice
func mapToSlice(productsMap map[string]models.Product) []models.Product {
	productsSlice := make([]models.Product, 0, len(productsMap))
	for _, product := range productsMap {
		productsSlice = append(productsSlice, product)
	}
	return productsSlice
}

// Function to process and insert products (SKUs) into the DB
func processProducts(db *gorm.DB, products []models.Product, regionID uint) {
	for _, product := range products {
		// Check and parse VCPU, default to 0 if missing
		vcpu, _ := strconv.Atoi(utils.DefaultIfEmpty(product.Attributes["vcpu"], "0"))

		// Create SKU record
		sku := models.SKU{
			SKUCode:         product.SKU,
			RegionID:        regionID,
			InstanceSKU:     product.Attributes["instancesku"],
			ProductFamily:   product.ProductFamily,
			VCPU:            vcpu,
			Type:            product.Attributes["usagetype"],
			OperatingSystem: product.Attributes["operatingSystem"],
			InstanceType:    product.Attributes["instanceType"],
			Storage:         product.Attributes["storage"],
			Network:         product.Attributes["networkPerformance"],
			CpuArchitecture: product.Attributes["processorArchitecture"],
			Memory: 		 product.Attributes["memory"],
		}

		// Insert SKU (check if it exists, create if not)
		if err := db.FirstOrCreate(&sku, models.SKU{SKUCode: sku.SKUCode}).Error; err != nil {
			log.Printf("Failed to insert SKU %s: %v", product.SKU, err)
		} else {
			log.Printf("Successfully inserted SKU: %s", product.SKU)
		}
	}
}
func processTerms(db *gorm.DB, terms map[string]map[string]models.TermDetails) {
	for skuCode, termData := range terms {
		log.Printf("Processing Term Sku: %s", skuCode)

		for termType, termDetails := range termData {
			log.Printf("Processing TermType: %s", termType)

			// Fetch the SKU_ID for the given SKU code
			var skuID uint
			if err := db.Table("skus").Select("id").Where("sku_code = ?", skuCode).Scan(&skuID).Error; err != nil {
				log.Printf("Failed to find SKU_ID for SKU %s: %v", skuCode, err)
				continue
			}

			// Extract the PriceDimension data
			for priceKey, priceDetails := range termDetails.PriceDimensions {
				log.Printf("Processing PriceDimension: %s for SKU: %s", priceKey, skuCode)
				pricePerUnit := priceDetails.PricePerUnit["USD"]

				// Create a term entry in Price
				termEntry := models.Price{
					SKU_ID:        skuID, // Ensure the type matches (uint)
					EffectiveDate: termDetails.EffectiveDate,
					Unit:          priceDetails.Unit,
					PricePerUnit:  pricePerUnit,
				}

				// Insert the term entry into the database
				if err := db.Create(&termEntry).Error; err != nil {
					log.Printf("Failed to insert term for SKU %s: %v", skuCode, err)
					continue
				}

				log.Printf("Successfully inserted term for SKU %s", skuCode)

				// Check if TermAttributes have non-empty values
				leaseContractLength := termDetails.TermAttributes.LeaseContractLength
				purchaseOption := termDetails.TermAttributes.PurchaseOption
				offeringClass := termDetails.TermAttributes.OfferingClass

				if leaseContractLength != "" || purchaseOption != "" || offeringClass != "" {
					// Insert term attributes only if there are non-empty values
					termAttributes := models.Term{
						SKU_ID:              skuID,
						LeaseContractLength: leaseContractLength,
						PurchaseOption:      purchaseOption,
						OfferingClass:       offeringClass,
						PriceID:             termEntry.PriceID,
					}

					// Insert term attributes into the database
					if err := db.Create(&termAttributes).Error; err != nil {
						log.Printf("Failed to insert termAttributes for SKU %s: %v", skuCode, err)
						continue
					}

					log.Printf("Successfully inserted termAttributes for SKU %s", skuCode)
				} else {
					log.Printf("No valid TermAttributes for SKU %s; skipping insertion.", skuCode)
				}
			}
		}
	}
}