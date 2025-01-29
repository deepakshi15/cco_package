package track
import (
	"encoding/json"
	"project/models"
	"os"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
)
func CreateEmptyTrackFile(fileName string) {
	initialState := models.RegionState{
		RegionName: "",
		State:      "processed",
	}
	writeTrackFile(fileName, initialState)
}
func ReadTrackFile(fileName string) models.RegionState {
	var state models.RegionState
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading track file: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}
	return state
}
func UpdateTrackFile(fileName, regionName, state string) {
	newState := models.RegionState{
		RegionName: regionName,
		State:      state,
	}
	writeTrackFile(fileName, newState)
}
func writeTrackFile(fileName string, state models.RegionState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		fmt.Printf("Error writing to track file: %v\n", err)
		os.Exit(1)
	}
}
func RemoveRegionData(db *gorm.DB, regionName string) {
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
			return
		}
		tx.Rollback()
		fmt.Printf("Failed to fetch region: %v\n", err)
		return
	}

	// Delete the associated SKU records first
	if err := tx.Where("region_id = ?", region.RegionID).Delete(&models.SKU{}).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to delete SKU data: %v\n", err)
		return
	}

	// Optionally delete other related data, like SavingsPlan
	if err := tx.Where("region_id = ?", region.RegionID).Delete(&models.SavingPlan{}).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to delete savings plan data: %v\n", err)
		return
	}

	// Delete the region itself
	if err := tx.Delete(&region).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to delete region: %v\n", err)
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return
	}

	fmt.Printf("Successfully removed data for region: %s\n", regionName)
}