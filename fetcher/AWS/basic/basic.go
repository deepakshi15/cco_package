package basic

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"gorm.io/gorm"
	"cco-package/fetcher/AWS/models"
    "cco-package/fetcher/AWS/utils"
	"fmt"
)
func ProcessCurrentVersionFile(db *gorm.DB, filepath string, regionID uint) error {
    file, err := os.Open(filepath)
    if err != nil {
        return fmt.Errorf("failed to open current version file: %v", err)
    }
    defer file.Close()

    var data models.PricingData

    err = json.NewDecoder(file).Decode(&data)
    if err != nil {
        return fmt.Errorf("failed to decode current version file: %v", err)
    }

    // Convert map to slice and call function to process products (SKUs)
    productsSlice := mapToSlice(data.Products)
    err = processProducts(db, productsSlice, regionID)
    if err != nil {
        return fmt.Errorf("failed to process products: %v", err)
    }

    // Call function to process terms
    err = processTerms(db, data.Terms["OnDemand"])
    if err != nil {
        return fmt.Errorf("failed to process on-demand terms: %v", err)
    }

    err = processTerms(db, data.Terms["Reserved"])
    if err != nil {
        return fmt.Errorf("failed to process reserved terms: %v", err)
    }

    return nil
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
func processProducts(db *gorm.DB, products []models.Product, regionID uint) error {
	for _, product := range products {
		// Check and parse VCPU, default to 0 if missing
		vcpu, err := strconv.Atoi(utils.DefaultIfEmpty(product.Attributes["vcpu"], "0"))
		if err != nil {
			return fmt.Errorf("failed to convert vcpu for SKU %s: %v", product.SKU, err)
		}

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
			Memory:          product.Attributes["memory"],
		}

		// Insert SKU (check if it exists, create if not)
		if err := db.FirstOrCreate(&sku, models.SKU{SKUCode: sku.SKUCode}).Error; err != nil {
			return fmt.Errorf("failed to insert SKU %s: %v", product.SKU, err)
		} else {
			log.Printf("Successfully inserted SKU: %s", product.SKU)
		}
	}

	return nil
}

func processTerms(db *gorm.DB, terms map[string]map[string]models.TermDetails) error {
	for skuCode, termData := range terms {
		log.Printf("Processing Term SKU: %s", skuCode)

		for termType, termDetails := range termData {
			log.Printf("Processing TermType: %s", termType)

			// Fetch the SKU_ID for the given SKU code
			var skuID uint
			if err := db.Table("skus").Select("id").Where("sku_code = ?", skuCode).Scan(&skuID).Error; err != nil {
				return fmt.Errorf("failed to find SKU_ID for SKU %s: %v", skuCode, err)
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
					return fmt.Errorf("failed to insert term for SKU %s: %v", skuCode, err)
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
						return fmt.Errorf("failed to insert termAttributes for SKU %s: %v", skuCode, err)
					}

					log.Printf("Successfully inserted termAttributes for SKU %s", skuCode)
				} else {
					log.Printf("No valid TermAttributes for SKU %s; skipping insertion.", skuCode)
				}
			}
		}
	}
	return nil
}
