package basic

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gorm.io/gorm"
	"cco-package/fetcher/AWS/models"
	"cco-package/fetcher/AWS/utils"
	"cco-package/fetcher/AWS/convertData"
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
	var regionCode string
	if err := db.Table("regions").Select("region_code").Where("region_id = ?", regionID).Scan(&regionCode).Error; err != nil || regionCode == "" {
		return fmt.Errorf("failed to fetch region_code for regionID %d: %v", regionID, err)
	}
	fmt.Printf("Fetched regionCode: %s\n", regionCode)

	// Fetch the Provider ID for AWS
	var providerID uint
	if err := db.Table("providers").Select("provider_id").Where("provider_name = ?", "AWS").Scan(&providerID).Error; err != nil || providerID == 0 {
		return fmt.Errorf("failed to fetch provider ID for AWS: %v", err)
	}
	fmt.Printf("Fetched ProviderID: %d\n", providerID)

	for _, product := range products {
		// Check and parse VCPU, default to 0 if missing
		vcpu, err := strconv.Atoi(utils.DefaultIfEmpty(product.Attributes["vcpu"], "0"))
		if err != nil {
			return fmt.Errorf("failed to convert vcpu for SKU %s: %v", product.SKU, err)
		}
		if product.ProductFamily == "Compute Instance" || product.ProductFamily == "Compute Instance (bare metal)" {
			product.ProductFamily = "Compute"
		}

		// Convert and normalize fields
		networkData := convertData.ConvertNetwork(product.Attributes["networkPerformance"])
		memoryData := convertData.ConvertMemory(product.Attributes["memory"])

		// Extract additional attributes
		armSkuName := product.Attributes["armSkuName"]
		physicalProcessor := product.Attributes["physicalProcessor"]
		maxThroughput := product.Attributes["dedicatedEbsThroughput"]
		enhancedNetworking := product.Attributes["enhancedNetworkingSupported"]
		gpu := product.Attributes["gpuMemory"]
		maxIOPS := product.Attributes["maxIopsvolume"]

		// Create SKU record
		sku := models.SKU{
			SKUCode:             product.SKU,
			RegionID:            regionID,
			ProviderID:          providerID,
			RegionCode:          regionCode,
			ArmSkuName:          armSkuName,
			InstanceSKU:         product.Attributes["instancesku"],
			ProductFamily:       product.ProductFamily,
			VCPU:                vcpu,
			Type:                product.Attributes["usagetype"],
			OperatingSystem:     product.Attributes["operatingSystem"],
			InstanceType:        product.Attributes["instanceType"],
			Storage:             product.Attributes["storage"],
			Network:             networkData,
			CpuArchitecture:     product.Attributes["processorArchitecture"],
			Memory:              memoryData,
			PhysicalProcessor:   physicalProcessor,
			MaxThroughput:       maxThroughput,
			EnhancedNetworking:  enhancedNetworking,
			GPU:                 gpu,
			MaxIOPS:             maxIOPS,
		}

		// Insert SKU (check if it exists, create if not)
		if err := db.FirstOrCreate(&sku, models.SKU{SKUCode: sku.SKUCode}).Error; err != nil {
			return fmt.Errorf("failed to insert SKU %s: %v", product.SKU, err)
		}
	}

	return nil
}

func processTerms(db *gorm.DB, terms map[string]map[string]models.TermDetails) error {
	for skuCode, termData := range terms {
		for _, termDetails := range termData {
			// Fetch the SKU_ID for the given SKU code
			var skuID uint
			if err := db.Table("skus").Select("id").Where("sku_code = ?", skuCode).Scan(&skuID).Error; err != nil {
				return fmt.Errorf("failed to find SKU_ID for SKU %s: %v", skuCode, err)
			}

			// Extract the PriceDimension data
			for _, priceDetails := range termDetails.PriceDimensions {
				pricePerUnit := priceDetails.PricePerUnit["USD"]

				// Create a term entry in Price
				termEntry := models.Price{
					SKU_ID:        skuID,
					EffectiveDate: termDetails.EffectiveDate,
					Unit:          priceDetails.Unit,
					PricePerUnit:  pricePerUnit,
				}

				// Insert the term entry into the database
				if err := db.Create(&termEntry).Error; err != nil {
					return fmt.Errorf("failed to insert term for SKU %s: %v", skuCode, err)
				}

				// Check if TermAttributes have non-empty values
				leaseContractLength := termDetails.TermAttributes.LeaseContractLength
				purchaseOption := termDetails.TermAttributes.PurchaseOption
				offeringClass := termDetails.TermAttributes.OfferingClass

				if leaseContractLength != "" || purchaseOption != "" || offeringClass != "" {
					// Insert term attributes only if there are non-empty values
					termAttributes := models.Term{
						SKU_ID:              skuID,
						LeaseContractLength: convertData.ConvertYear(leaseContractLength),
						PurchaseOption:      purchaseOption,
						OfferingClass:       offeringClass,
						PriceID:             termEntry.PriceID,
					}

					// Insert term attributes into the database
					if err := db.Create(&termAttributes).Error; err != nil {
						return fmt.Errorf("failed to insert termAttributes for SKU %s: %v", skuCode, err)
					}
				}
			}
		}
	}
	return nil
}