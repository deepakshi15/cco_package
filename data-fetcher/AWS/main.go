package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"project/basic"
	"project/config"
	"project/models"
	"project/saving"
	"project/utils"
	"project/track"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Step 1: Initialize the Database Connection
	db, err := gorm.Open(postgres.Open(config.DbConnStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	
	// Step 2: Auto-Migrate the Tables (Including SavingPlan)
	db.AutoMigrate(&models.Provider{}, &models.Region{}, &models.SKU{}, &models.Price{}, &models.Term{}, &models.SavingPlan{})
	
		trackFile := "track.json"
		if _, err := os.Stat(trackFile); os.IsNotExist(err) {
			// If file doesn't exist, create an empty one
			track.CreateEmptyTrackFile(trackFile)
		}
	
		// Step 2: Read the current state from track.json
		state := track.ReadTrackFile(trackFile)
		if state.State != "processed" {
			fmt.Println("Previous region is not processed. Cleaning data...")
			// Call your function to remove data related to the region
			track.RemoveRegionData(db, state.RegionName)
		}
	 
	
		
		// Step 3: Create the Folder for the PriceList
		err = os.MkdirAll(config.PriceListPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create price-list directory: %v", err)
		}
	
		// Step 4: Initialize Provider and Service
		provider := models.Provider{ProviderName: "AWS"}
		db.FirstOrCreate(&provider, models.Provider{ProviderName: "AWS"})
	
		// Step 5: Download the region_index.json file (Basic Plan)
		regionFilePath := filepath.Join(config.PriceListPath, "region_index.json")
		err = utils.DownloadFile(config.RegionURL, regionFilePath)
		if err != nil {
			log.Fatalf("Failed to download region: %v", err)
		}
	
		// Step 6: Download the saving_region_index.json file (Saving Plan)
		savingRegionFilePath := filepath.Join(config.PriceListPath, "saving_region_index.json")
		err = utils.DownloadFile(config.SavingRegionURL, savingRegionFilePath)
		if err != nil {
			log.Fatalf("Failed to download saving region: %v", err)
		}
	
		// Step 7: Open and Process the region_index.json file (Basic Plan)
		regionFile, err := os.Open(regionFilePath)
		if err != nil {
			log.Fatalf("Failed to open region file: %v", err)
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
			log.Fatalf("Failed to decode region index file: %v", err)
		}
	
		// Step 8: Open and Process the saving_region_index.json file (Saving Plan)
		savingRegionFile, err := os.Open(savingRegionFilePath)
		if err != nil {
			log.Fatalf("Failed to open saving region file: %v", err)
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
			log.Fatalf("Failed to decode saving region index file: %v", err)
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
			db.FirstOrCreate(&regionEntry, models.Region{RegionCode: region.RegionCode})
	
			// Download the current version file for the region (Basic Plan)
			currentVersionURL := config.BaseURL + region.CurrentVersionUrl
			currentVersionFile := filepath.Join(config.PriceListPath, region.RegionCode+".json")
			err = utils.DownloadFile(currentVersionURL, currentVersionFile)
			if err != nil {
				log.Printf("Failed to download %s: %v", currentVersionURL, err)
				continue
			}
	
			// Process the Basic Plan current version file
			basic.ProcessCurrentVersionFile(db, currentVersionFile, regionEntry.RegionID)
	
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
					log.Printf("Processing Saving Plan for region: %s", savingRegion.RegionCode)
	
					// Download the saving plan version file for the region
					savingVersionURL := config.BaseURL + savingRegion.VersionUrl
					savingVersionFile := filepath.Join(config.PriceListPath, savingRegion.RegionCode+".json")
					err = utils.DownloadFile(savingVersionURL, savingVersionFile)
					if err != nil {
						log.Printf("Failed to download saving version file %s: %v", savingVersionURL, err)
						continue
					}
	
					// Process the Saving Plan version file
					saving.ProcessVersionFile(db, savingVersionFile, regionEntry.RegionID)
	
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
	}