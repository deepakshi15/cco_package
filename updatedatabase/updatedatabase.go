package updatedatabase

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var mainDb *gorm.DB
var tempDb *gorm.DB
var backupDb *gorm.DB

var tables = []string{
	"providers",
	"services",
	"regions",
	"skus",
	"prices",
	"terms",
	"saving_plans",
}

// Connects to databases and ensures global variables are updated
func connectToDatabases() error {
	var err error

	mainDb, err = gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=main_db port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to main DB: %w", err)
	}

	tempDb, err = gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=temp_db port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to temp DB: %w", err)
	}

	backupDb, err = gorm.Open(postgres.Open("host=localhost user=postgres password=password dbname=backup_db port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to backup DB: %w", err)
	}

	return nil
}

// Main update function
func Updatedatabase() error {
	if err := connectToDatabases(); err != nil {
		return err
	}

	var count int64
	if err := mainDb.Table("providers").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if main_db has data: %w", err)
	}

	// If main DB has data, move it to backup
	if count > 0 {
		fmt.Println("Transferring data from main_db to backup_db...")
		if err := transferData(mainDb, backupDb); err != nil {
			return fmt.Errorf("failed to transfer data to backup_db: %w", err)
		}

		fmt.Println("Emptying main_db...")
		if err := emptyDatabase(mainDb); err != nil {
			return fmt.Errorf("failed to empty main_db: %w", err)
		}
	}

	// Move temp_db data to main_db
	fmt.Println("Inserting data from temp_db to main_db...")
	if err := transferData(tempDb, mainDb); err != nil {
		return fmt.Errorf("failed to insert temp_db data into main_db: %w", err)
	}

	// Empty temp_db
	fmt.Println("Emptying temp_db...")
	if err := emptyDatabase(tempDb); err != nil {
		return fmt.Errorf("failed to empty temp_db: %w", err)
	}

	fmt.Println("Database migration completed successfully!")
	return nil
}

// Transfers data from sourceDb to targetDb
func transferData(sourceDb, targetDb *gorm.DB) error {
	const batchSize = 1000

	if err := emptyDatabase(targetDb); err != nil {
		return fmt.Errorf("failed to empty targetDb before data transfer: %w", err)
	}

	for _, table := range tables {
		fmt.Printf("Transferring data for table: %s...\n", table)

		var records []map[string]interface{}
		if err := sourceDb.Table(table).Find(&records).Error; err != nil {
			return fmt.Errorf("failed to fetch data from table %s: %w", table, err)
		}

		for i := 0; i < len(records); i += batchSize {
			end := i + batchSize
			if end > len(records) {
				end = len(records)
			}

			batch := records[i:end]
			if table == "prices" {
				batch = filterValidPrices(batch, targetDb)
			}

			if len(batch) > 0 {
				if err := targetDb.Table(table).Create(&batch).Error; err != nil {
					return fmt.Errorf("failed to insert data into table %s (batch %d-%d): %w", table, i, end, err)
				}
			}
		}
	}
	return nil
}

// Filters out invalid price entries
func filterValidPrices(batch []map[string]interface{}, targetDb *gorm.DB) []map[string]interface{} {
	var validBatch []map[string]interface{}

	for _, record := range batch {
		if skuID, ok := record["sku_id"]; ok {
			var count int64
			if err := targetDb.Table("skus").Where("id = ?", skuID).Count(&count).Error; err != nil || count == 0 {
				fmt.Printf("Skipping record with invalid SKU_ID: %v\n", skuID)
				continue
			}
		}
		validBatch = append(validBatch, record)
	}
	return validBatch
}

// Empties all tables in the given database except backup
func emptyDatabase(db *gorm.DB) error {
	if db == backupDb {
		fmt.Println("Skipping emptying backup_db")
		return nil
	}

	for _, table := range tables {
		fmt.Printf("Emptying table: %s...\n", table)
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)).Error; err != nil {
			return fmt.Errorf("failed to empty table %s: %w", table, err)
		}
	}
	return nil
}
