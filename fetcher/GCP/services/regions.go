package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gorm.io/gorm"

	"cco-package/fetcher/GCP/config"
	"cco-package/fetcher/GCP/models"
)


func FetchAndStoreRegions() error {
	// Step 1: Check or insert GCP provider
	var provider models.Provider
	err := config.DB.Where("provider_name = ?", "GCP").First(&provider).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		provider = models.Provider{
			ProviderName: "GCP",
			CreatedDate:  time.Now(),
			ModifiedDate: time.Now(),
			DisableFlag:  false,
		}
		if err := config.DB.Create(&provider).Error; err != nil {
			return fmt.Errorf("failed to insert provider: %w", err)
		}
		fmt.Println("Inserted new provider: GCP")
	} else if err != nil {
		return fmt.Errorf("error checking provider: %w", err)
	} else {
		fmt.Println("Provider GCP already exists")
	}

	// Step 2: Call GCP API for regions
	regionList, err := getGCPRegions()
	if err != nil {
		return fmt.Errorf("failed to fetch regions: %w", err)
	}

	// Step 3: Insert or update regions
	for _, item := range regionList.Items {
		var existing models.Region
		err := config.DB.Where("region_code = ?", item.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newRegion := models.Region{
				RegionCode:   item.Name,
				ProviderID:   provider.ProviderID,
				CreatedDate:  time.Now(),
				ModifiedDate: time.Now(),
				DisableFlag:  false,
			}
			if err := config.DB.Create(&newRegion).Error; err != nil {
				fmt.Printf("Failed to insert region %s: %v\n", item.Name, err)
			} else {
				fmt.Printf("Inserted region: %s\n", item.Name)
			}
		} else if err != nil {
			fmt.Printf("Error checking region %s: %v\n", item.Name, err)
		} else {
			fmt.Printf("Region already exists: %s\n", item.Name)
		}
	}

	return nil
}

func getGCPRegions() (*models.RegionList, error) {
	url := "https://www.googleapis.com/compute/v1/projects/177423693614/regions"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", config.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("Response Status Code:", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var regionList models.RegionList
	if err := json.Unmarshal(body, &regionList); err != nil {
		fmt.Println("Raw response:", string(body)) // for debugging
		return nil, err
	}

	return &regionList, nil
}
