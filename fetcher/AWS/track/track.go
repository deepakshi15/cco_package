package track

import (
	"encoding/json"
	"cco-package/fetcher/AWS/models"
	"fmt"
	"gorm.io/gorm"
	"os"
)

func CreateEmptyTrackFile(fileName string) error {
	initialState := models.RegionState{
		RegionName: "",
		State:      "processed",
	}
	return writeTrackFile(fileName, initialState)
}

func ReadTrackFile(fileName string) (models.RegionState, error) {
	var state models.RegionState
	data, err := os.ReadFile(fileName)
	if err != nil {
		return state, fmt.Errorf("error reading track file: %v", err)
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		return state, fmt.Errorf("error parsing JSON: %v", err)
	}
	return state, nil
}

func UpdateTrackFile(fileName, regionName, state string) error {
	newState := models.RegionState{
		RegionName: regionName,
		State:      state,
	}
	return writeTrackFile(fileName, newState)
}

func writeTrackFile(fileName string, state models.RegionState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to track file: %v", err)
	}
	return nil
}

func RemoveRegionData(db *gorm.DB, regionName string) error {
	fmt.Printf("Removing data for region: %s\n", regionName)

	// Begin a transaction for atomicity
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()

	// Find the region by name
	var region models.Region
	if err := tx.Where("region_code = ?", regionName).First(&region).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("Region not found: %s\n", regionName)
			return nil
		}
		tx.Rollback()
		return fmt.Errorf("failed to fetch region: %v", err)
	}

	// Delete the associated SKU records first
	if err := tx.Where("region_id = ?", region.RegionID).Delete(&models.SKU{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete SKU data: %v", err)
	}

	// Optionally delete other related data, like SavingsPlan
	if err := tx.Where("region_id = ?", region.RegionID).Delete(&models.SavingPlan{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete savings plan data: %v", err)
	}

	// Delete the region itself
	if err := tx.Delete(&region).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete region: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Printf("Successfully removed data for region: %s\n", regionName)
	return nil
}
