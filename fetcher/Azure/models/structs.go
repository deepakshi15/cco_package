package models

import "time"

type Provider struct {
	ProviderID   uint      `gorm:"primaryKey;autoIncrement"`
	ProviderName string    `gorm:"size:50;not null"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

func (Provider) TableName() string {
	return "providers" // Explicitly specify the table name
}

type Region struct {
	RegionID    uint      `gorm:"primaryKey;autoIncrement"`
	ProviderID  uint      `gorm:"not null"`
	RegionCode  string    `gorm:"size:20;not null"`
	CreatedDate time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag bool      `gorm:"default:false"`
}

func (Region) TableName() string {
	return "regions" // Explicitly specify the table name
}

type SKU struct {
	ID              uint   `gorm:"primaryKey"`
	RegionID        uint   `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	RegionCode      string `gorm:"not null"`
	SKUCode         string `gorm:"unique"`
	ArmSkuName      string `gorm:"column:arm_sku_name"`
	Name            string
	InstanceSKU     string
	ProductFamily   string
	VCPU            int
	CpuArchitecture string
	InstanceType    string
	Storage         string
	Network         string
	OperatingSystem string
	Type 		    string 
	Memory          string
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

func (SKU) TableName() string {
    return "skus"
}

// Term represents the terms table
type Term struct {
	OfferTermID         uint       `gorm:"primaryKey"`
	PriceID             uint       `gorm:"not null"`
	SkuID               uint       `gorm:"not null"`  // Changed to uint to match the type of sku.ID
	PurchaseOption      *string    `gorm:"size:100"`
	LeaseContractLength *string    `gorm:"size:50"`
	OfferingClass       *string    `gorm:"size:50"`
	CreatedDate         time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	ModifiedDate        time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	DisableFlag         bool       `gorm:"default:false"`
}

// TableName specifies the table name for Term
func (Term) TableName() string {
	return "terms"
}

type Price struct {
	PriceID       int       `gorm:"primaryKey;autoIncrement"`    // Primary Key, Auto-incremented
	SkuID         uint       `gorm:"not null"`                    // Foreign key referencing sku table
	PricePerUnit  float64   `gorm:"type:numeric(15,6)"` // Price per unit (numeric field with precision)
	Unit          string    `gorm:"size:255;not null"`           // Unit of measurement
	EffectiveDate time.Time `gorm:"not null"`                    // Effective date for the price
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

// TableName specifies the table name for Price
func (Price) TableName() string {
	return "prices"
}