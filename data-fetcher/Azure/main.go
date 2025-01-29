package main

import (
	"log"
	"cco_backend/config"
	"cco_backend/services"
)

func main() {
	// Initialize the database
	config.ConnectDatabase()

	// Import data from Azure VM API
	// if err := services.ImportData(); err != nil {
	// 	log.Fatalf("Error importing Azure VM data: %v", err)
	// } else {
	// 	log.Println("Azure VM data import completed successfully.")
	// }

	// Import SKU data
	// if err := services.ImportSkuData(); err != nil { 
	// 	log.Fatalf("Error importing SKU data: %v", err)
	// } else {
	// 	log.Println("SKU data import completed successfully.")
	// }

	//Import terms data
	if err := services.ImportTermsData(); err != nil {
		log.Fatalf("Error importing terms data: %v", err)
	}

	// Import prices data
	// if err := services.ImportPricesData(); err != nil {
	// 	log.Fatalf("Error importing prices data: %v", err)
	// } else {
	// 	log.Println("Prices data import completed successfully.")
	// }
}