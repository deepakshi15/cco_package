package AWS

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	
	"cco-package/fetcher/AWS/models"
	"cco-package/fetcher/AWS/saving"
	"cco-package/fetcher/AWS/track"
	"cco-package/fetcher/AWS/utils"
	"cco-package/fetcher/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"cco-package/fetcher/AWS/basic"
)

func RunAWS() error {
	// Step 1: Initialize the Database Connection (Using the global DB in config)
	var err error
	config.DB, err = gorm.Open(postgres.Open(config.DbConnStr), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Step 2: Auto-Migrate thae Tables (Including SavingPlan)
	err = config.DB.AutoMigrate(&models.Provider{}, &models.Region{}, &models.SKU{}, &models.Price{}, &models.Term{}, &models.SavingPlan{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %v", err)
	}

	// Step 3: Track File Handling
	trackFile := "track.json"
	if _, err := os.Stat(trackFile); os.IsNotExist(err) {
		// If file doesn't exist, create an empty one
		err = track.CreateEmptyTrackFile(trackFile)
		if err != nil {
			return fmt.Errorf("failed to create empty track file: %v", err)
		}
	}

	// Step 4: Read the current state from track.json
	state, err := track.ReadTrackFile("track.json")
	if err != nil {
		return fmt.Errorf("failed to read track file: %v", err)
	}
	
	if state.State != "processed" {
		fmt.Println("Previous region is not processed. Cleaning data...")
		err = track.RemoveRegionData(config.DB, state.RegionName)
		if err != nil {
			return fmt.Errorf("failed to remove region data: %v", err)
		}
	}
	

	// Step 5: Create the Folder for the PriceList
	err = os.MkdirAll(config.PriceListPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create price-list directory: %v", err)
	}

	// Step 6: Initialize Provider and Service
	provider := models.Provider{ProviderName: "AWS"}
	config.DB.FirstOrCreate(&provider, models.Provider{ProviderName: "AWS"})

	// Step 7: Download the region_index.json file (Basic Plan)
	regionFilePath := filepath.Join(config.PriceListPath, "region_index.json")
	err = utils.DownloadFile(config.RegionURL, regionFilePath)
	if err != nil {
		return fmt.Errorf("failed to download region: %v", err)
	}

	// Step 8: Download the saving_region_index.json file (Saving Plan)
	savingRegionFilePath := filepath.Join(config.PriceListPath, "saving_region_index.json")
	err = utils.DownloadFile(config.SavingRegionURL, savingRegionFilePath)
	if err != nil {
		return fmt.Errorf("failed to download saving region: %v", err)
	}

	// Step 9: Open and Process the region_index.json file (Basic Plan)
	regionFile, err := os.Open(regionFilePath)
	if err != nil {
		return fmt.Errorf("failed to open region file: %v", err)
	}
	defer regionFile.Close()

	var regionData struct {
		Regions map[string]struct {
			RegionCode        string `json:"regionCode"`
			CurrentVersionUrl string `json:"currentVersionUrl"`
		} `json:"regions"`
	}
	err = json.NewDecoder(regionFile).Decode(&regionData)
	if err != nil {
		return fmt.Errorf("failed to decode region index file: %v", err)
	}

	// Step 10: Open and Process the saving_region_index.json file (Saving Plan)
	savingRegionFile, err := os.Open(savingRegionFilePath)
	if err != nil {
		return fmt.Errorf("failed to open saving region file: %v", err)
	}
	defer savingRegionFile.Close()

	// Updated structure for saving region data
	var savingRegionData struct {
		Regions []struct {
			RegionCode string `json:"regionCode"`
			VersionUrl string `json:"versionUrl"`
		} `json:"regions"`
	}

	// Decode the saving region index file
	err = json.NewDecoder(savingRegionFile).Decode(&savingRegionData)
	if err != nil {
		return fmt.Errorf("failed to decode saving region index file: %v", err)
	}

	// Process each region for Basic Plan and Saving Plan
	for regionCode, region := range regionData.Regions {
		log.Printf("Processing region: %s", region.RegionCode)
		track.UpdateTrackFile(trackFile, region.RegionCode, "processing")
		// Insert the Region data into DB (for Basic Plan)
		regionEntry := models.Region{
			RegionCode: region.RegionCode,
			ProviderID: provider.ProviderID,
		}
		err = config.DB.FirstOrCreate(&regionEntry, models.Region{RegionCode: region.RegionCode}).Error
		if err != nil {
			return fmt.Errorf("failed to insert region data into DB: %v", err)
		}

		// Download the current version file for the region (Basic Plan)
		currentVersionURL := config.BaseURL + region.CurrentVersionUrl
		currentVersionFile := filepath.Join(config.PriceListPath, region.RegionCode+".json")
		err = utils.DownloadFile(currentVersionURL, currentVersionFile)
		if err != nil {
			log.Printf("Failed to download %s: %v", currentVersionURL, err)
			continue
		}

		// Process the Basic Plan current version file
		
		err = basic.ProcessCurrentVersionFile(config.DB, currentVersionFile, regionEntry.RegionID)
		if err != nil {
			return fmt.Errorf("failed to process current version file: %v", err)
		}

		// Remove the current version file (Basic Plan)
		err = os.Remove(currentVersionFile)
		if err != nil {
			log.Printf("Failed to delete file %s: %v", currentVersionFile, err)
		} else {
			log.Printf("Successfully deleted file: %s", currentVersionFile)
		}

		// Process the corresponding region in the saving region index (Saving Plan)
		for _, savingRegion := range savingRegionData.Regions {
			if savingRegion.RegionCode == regionCode {
			

				// Download the saving plan version file for the region
				savingVersionURL := config.BaseURL + savingRegion.VersionUrl
				savingVersionFile := filepath.Join(config.PriceListPath, savingRegion.RegionCode+".json")
				err = utils.DownloadFile(savingVersionURL, savingVersionFile)
				if err != nil {
					log.Printf("Failed to download saving version file %s: %v", savingVersionURL, err)
					continue
				}

				// Process the Saving Plan version file
				err = saving.ProcessVersionFile(config.DB, savingVersionFile, regionEntry.RegionID)
				if err != nil {
					return fmt.Errorf("failed to process saving version file: %v", err)
				}

				// Remove the saving plan version file
				err = os.Remove(savingVersionFile)
				if err != nil {
					log.Printf("Failed to delete saving version file %s: %v", savingVersionFile, err)
				} else {
					track.UpdateTrackFile(trackFile, savingRegion.RegionCode, "processed")
					log.Printf("Successfully deleted saving version file: %s", savingVersionFile)
				}
			}
		}
	}

	log.Println("Processing complete.")
	return nil
}

