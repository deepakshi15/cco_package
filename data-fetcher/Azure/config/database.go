package config

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres" //PostgreSQL driver for GORM, used to interact with PostgreSQL databases
	"gorm.io/gorm"            //core GORM package that provides the ORM functionality

	"data-fetcher/Azure/models" // Import your models package
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=localhost user=postgres password=password dbname=cloudcost port=5432 sslmode=disable" // Database connection string
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Database connected successfully!")

	// Automigrate your models here
	err = DB.AutoMigrate(
		&models.Provider{},
		&models.Region{},
		&models.SKU{},   // Your Sku model
		&models.Term{},  // Your Term model
		&models.Price{}, // Your Price model (add all relevant models here)
	)
	if err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}
	fmt.Println("Database migration completed successfully!")
}
