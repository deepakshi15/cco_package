package models

import (
	"time"
)

type RegionState struct {
    RegionName string `json:"region_name"`
    State      string `json:"state"`
}

type Provider struct {
	ProviderID   uint   `gorm:"primaryKey"`
	ProviderName string `gorm:"unique"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

type Region struct {
	RegionID   uint   `gorm:"primaryKey"`
	RegionCode string `gorm:"unique"`
	ProviderID uint   `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

type SKU struct {
	ID              uint   `gorm:"primaryKey"`
	RegionID        uint `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
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

type Price struct {
	PriceID       uint      `gorm:"primaryKey;autoIncrement"`
	SKU_ID        uint      `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	EffectiveDate string    `gorm:"type:varchar(255)"`
	Unit          string    `gorm:"type:varchar(50)"`
	PricePerUnit  string    `gorm:"type:varchar(50)"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

type Term struct {
	OfferTermID         int    `gorm:"primaryKey;autoIncrement"`
	SKU_ID              uint   `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	PriceID             uint   `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	LeaseContractLength string `gorm:"size:255"`
	PurchaseOption      string `gorm:"size:255"`
	OfferingClass       string `gorm:"size:255"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

type SavingPlan struct {
	ID                  uint   `gorm:"primaryKey"`
	DiscountedSku       string
	Sku                 string
	LeaseContractLength int
	DiscountedRate      string
	RegionID            uint `gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key with cascade delete
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}


// ! Iska usage kya hai pata nhi filhal
type JSON map[string]interface{}
type PricingData struct {
	Products map[string]Product                           `json:"products"`
	Terms    map[string]map[string]map[string]TermDetails `json:"terms"`
}

type Product struct {
	SKU           string            `json:"sku"`
	ProductFamily string            `json:"productFamily"`
	Attributes    map[string]string `json:"attributes"`
}
type PricePerUnit struct {
	USD string `json:"USD"`
}
type PriceDimension struct {
	RateCode     string            `json:"rateCode"`
	Description  string            `json:"description"`
	BeginRange   string            `json:"beginRange"`
	EndRange     string            `json:"endRange"`
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
	AppliesTo    []string          `json:"appliesTo"`
}
type TermAttributes struct {
	LeaseContractLength string `json:"LeaseContractLength"`
	PurchaseOption      string `json:"PurchaseOption"`
	OfferingClass       string `json:"OfferingClass"`
}
type TermDetails struct {
	OfferTermCode   string                    `json:"offerTermCode"`
	Sku             string                    `json:"sku"`
	EffectiveDate   string                    `json:"effectiveDate"`
	PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
	TermAttributes  TermAttributes            `json:"termAttributes"` // This should be a struct, not a map
}

type SavingTermDetails struct {
	Sku                 string `json:"sku"`
	Description         string `json:"description"`
	EffectiveDate       string `json:"effectiveDate"`
	LeaseContractLength  struct {
		Duration int    `json:"duration"`
	} `json:"leaseContractLength"`
	Rates []RateDetails `json:"rates"`
}

type RateDetails struct {
	DiscountedSku         string `json:"discountedSku"`
	DiscountedUsageType   string `json:"discountedUsageType"`
	DiscountedOperation   string `json:"discountedOperation"`
	DiscountedServiceCode string `json:"discountedServiceCode"`
	RateCode              string `json:"rateCode"`
	Unit                  string `json:"unit"`
	DiscountedRate        struct {
		Price    string `json:"price"`
		Currency string `json:"currency"`
	} `json:"discountedRate"`
	DiscountedRegionCode   string `json:"discountedRegionCode"`
	DiscountedInstanceType string `json:"discountedInstanceType"`
}

type SavingData struct {
	TermsPlan struct {
		SavingsPlan []SavingTermDetails `json:"savingsPlan"`
	} `json:"terms"`
}
