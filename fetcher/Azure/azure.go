package Azure

import (
	"log"
	"cco-package/fetcher/config"
	"cco-package/fetcher/Azure/services"
)

// RunAzure fetches data from Azure and returns an error if any step fails.
func RunAzure() error {
	// Initialize the database
	if err := config.ConnectDatabase(); err != nil {
		return err
	}

	// Import data from Azure VM API
	if err := services.ImportData(); err != nil {
		log.Printf("Error importing Azure VM data: %v", err)
		return err
	}
	log.Println("Azure VM data import completed successfully.")

	// Import terms data
	if err := services.ImportTermsData(); err != nil {
		log.Printf("Error importing terms data: %v", err)
		return err
	}

	// Uncomment if needed
	// if err := services.ImportSkuData(); err != nil { 
	// 	log.Printf("Error importing SKU data: %v", err)
	// 	return err
	// }
	// log.Println("SKU data import completed successfully.")

	// if err := services.ImportPricesData(); err != nil {
	// 	log.Printf("Error importing prices data: %v", err)
	// 	return err
	// }
	// log.Println("Prices data import completed successfully.")

	return nil // No errors
}
