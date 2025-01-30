package services

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/models"
	"cco-package/fetcher/Azure/utils"
	"github.com/joho/godotenv"
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

	nextPageUrl := priceApiUrl
	pageCount := 0
	batchSize := 10

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

			skuCode, _ := safeString(priceItem["skuId"])
			productFamily, _ := safeString(priceItem["serviceFamily"])
			instanceSKU, ok := safeString(priceItem["armSkuName"])
			if !ok {
				log.Printf("Missing or invalid armSkuName: %v", priceItem)
				continue
			}
			usageType, ok := safeString(priceItem["type"])
			if !ok {
				log.Printf("Missing or invalid type: %v", priceItem)
				continue
			}
			regionName, _ := safeString(priceItem["armRegionName"])

			var matchedSku map[string]interface{}
			for _, skuItemInterface := range skuItems {
				skuItem, ok := skuItemInterface.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := safeString(skuItem["name"])
				if name == instanceSKU {
					matchedSku = skuItem
					break
				}
			}

			if matchedSku == nil {
				log.Printf("No matching SKU found for armSkuName: %s", instanceSKU)
				continue
			}

			name, _ := safeString(matchedSku["name"])
			var vCPU int
			var memory, cpuArchitecture, network, operatingSystem string

			capabilities, ok := matchedSku["capabilities"].([]interface{})
			if ok {
				for _, capabilityInterface := range capabilities {
					capability, ok := capabilityInterface.(map[string]interface{})
					if !ok {
						continue
					}
					switch capName, _ := safeString(capability["name"]); capName {
					case "vCPUs":
						vCPU = atoi(capability["value"].(string))
					case "MemoryGB":
						memory, _ = safeString(capability["value"])
					case "CpuArchitectureType":
						cpuArchitecture, _ = safeString(capability["value"])
					case "MaxNetworkInterfaces":
						network, _ = safeString(capability["value"])
					case "ProductName": // Corrected to fetch operating system data
						operatingSystem, _ = safeString(capability["value"])
					}
				}
			}

			region := models.Region{}
			if err := config.DB.Where("region_name = ?", regionName).First(&region).Error; err != nil {
				log.Printf("Error finding region: %v", err)
				continue
			}

			SKU := models.SKU{
				RegionID:        region.RegionID,
				InstanceSKU:     instanceSKU, // Correct mapping from armSkuName
				Name:            name,
				Type:            usageType,
				SKUCode:         skuCode,
				ProductFamily:   productFamily,
				VCPU:            vCPU,
				Memory:          memory,
				CpuArchitecture: cpuArchitecture,
				Network:         network,         // Corrected to map from MaxNetworkInterfaces
				OperatingSystem: operatingSystem, // Corrected to map from ProductName
				Storage:         "",
			}

			result := config.DB.Create(&SKU)
			if result.Error != nil {
				log.Printf("Error inserting SKU: %v", result.Error)
			} else {
				log.Printf("SKU inserted successfully: %v", SKU.Name)
			}
		}

		pageCount++
		if pageCount >= batchSize {
			log.Printf("Processed %d pages in this batch. Fetching next batch...", batchSize)
			pageCount = 0
			time.Sleep(2 * time.Second)
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