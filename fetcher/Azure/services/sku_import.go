package services

import (
	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/models"
	"cco-package/fetcher/Azure/utils"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func ImportSkuData() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID not found in environment variables")
	}

	priceApiUrl := "https://prices.azure.com/api/retail/prices?api-version=2023-01-01-preview&$filter=serviceName%20eq%20%27Virtual%20Machines%27"
	skuApiUrl := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Compute/skus?api-version=2024-07-01",
		subscriptionID,
	)

	bearerToken, err := utils.GenerateBearerToken()
	if err != nil {
		return fmt.Errorf("error generating bearer token: %w", err)
	}

	skuData, err := utils.FetchDataWithBearerToken(skuApiUrl, bearerToken)
	if err != nil {
		return fmt.Errorf("error fetching SKU data: %w", err)
	}

	skuItems, ok := skuData["value"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid format for SKU items")
	}

	// Fetch the Provider ID for Azure
	var providerID uint
	if err := config.DB.Table("providers").Select("provider_id").Where("provider_name = ?", "Azure").Scan(&providerID).Error; err != nil || providerID == 0 {
		return fmt.Errorf("failed to fetch provider ID for Azure: %v", err)
	}
	log.Printf("Fetched ProviderID: %d\n", providerID)

	nextPageUrl := priceApiUrl
	for nextPageUrl != "" {
		priceData, err := utils.FetchData(nextPageUrl)
		if err != nil {
			return fmt.Errorf("error fetching price data: %w", err)
		}

		priceItems, ok := priceData["Items"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid format for price items")
		}

		for _, priceItemInterface := range priceItems {
			priceItem, ok := priceItemInterface.(map[string]interface{})
			if !ok {
				log.Printf("Skipping invalid price item: %v", priceItemInterface)
				continue
			}

			skuID, _ := safeString(priceItem["skuId"])
			armSkuName, ok := safeString(priceItem["armSkuName"])
			if !ok || armSkuName == "" {
				log.Printf("Missing or invalid armSkuName: %v", priceItem)
				continue
			}
			skuType, _ := safeString(priceItem["type"])
			regionName, _ := safeString(priceItem["armRegionName"])

			var matchedSku map[string]interface{}
			for _, skuItemInterface := range skuItems {
				skuItem, ok := skuItemInterface.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := safeString(skuItem["name"])
				if name == armSkuName {
					matchedSku = skuItem
					break
				}
			}

			if matchedSku == nil {
				log.Printf("No matching SKU found for armSkuName: %s", armSkuName)
				continue
			}

			name, _ := safeString(matchedSku["name"])
			size, _ := safeString(matchedSku["size"])

			var vCPUs int
			var memoryGB, cpuArchitectureType, maxNetworkInterfaces string

			capabilities, ok := matchedSku["capabilities"].([]interface{})
			if ok {
				for _, capabilityInterface := range capabilities {
					capability, ok := capabilityInterface.(map[string]interface{})
					if !ok {
						continue
					}
					switch capName, _ := safeString(capability["name"]); capName {
					case "vCPUs":
						vCPUs = atoi(capability["value"].(string))
					case "MemoryGB":
						memoryGB, _ = safeString(capability["value"])
					case "CpuArchitectureType":
						cpuArchitectureType, _ = safeString(capability["value"])
					case "MaxNetworkInterfaces":
						maxNetworkInterfaces, _ = safeString(capability["value"])
					}
				}
			}

			region := models.Region{}
			if err := config.DB.Where("region_code = ?", regionName).First(&region).Error; err != nil {
				log.Printf("Error finding region: %v (region: %s)", err, regionName)
				continue
			}

			sku := models.SKU{
				RegionID:        region.RegionID,
				ProviderID:      providerID,
				RegionCode:      region.RegionCode,
				ArmSkuName:      armSkuName,
				Name:            name,
				Type:            skuType,
				SKUCode:         skuID,
				ProductFamily:   size,
				VCPU:            vCPUs,
				Memory:          memoryGB,
				CpuArchitecture: cpuArchitectureType,
				Network:         maxNetworkInterfaces,
			}

			result := config.DB.Create(&sku)
			if result.Error != nil {
				log.Printf("Error inserting SKU: %v", result.Error)
			} else {
				log.Printf("SKU inserted successfully: %v (armSkuName: %s)", sku.Name, sku.ArmSkuName)
			}
		}

		nextPageUrl, _ = safeString(priceData["NextPageLink"])
		log.Printf("Next page URL: %s", nextPageUrl)
	}

	log.Println("SKU data import completed successfully.")
	return nil
}

func safeString(value interface{}) (string, bool) {
	str, ok := value.(string)
	return str, ok
}

func atoi(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}