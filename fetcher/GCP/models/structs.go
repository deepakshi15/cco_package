package models

import "time"

// ========== Database Models ==========

// Provider DB model
type Provider struct {
	ProviderID   uint      `gorm:"primaryKey"`
	ProviderName string    `gorm:"unique"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

// Region DB model
type Region struct {
	RegionID     uint      `gorm:"primaryKey"`
	RegionCode   string    `gorm:"unique"`
	ProviderID   uint      `gorm:"not null;constraint:OnDelete:CASCADE;"`
	CreatedDate  time.Time `gorm:"default:current_timestamp"`
	ModifiedDate time.Time `gorm:"default:current_timestamp"`
	DisableFlag  bool      `gorm:"default:false"`
}

// SKU DB model
type SKU struct {
	ID                   uint      `gorm:"primaryKey"`
	RegionID             uint      `gorm:"not null;constraint:OnDelete:CASCADE;"`
	ProviderID           uint      `gorm:"not null"`
	RegionCode           string    `gorm:"not null"`
	SKUCode              string    `gorm:"unique"`
	ArmSkuName           string    `gorm:"column:arm_sku_name"`
	InstanceSKU          string
	ProductFamily        string
	VCPU                 int
	CpuArchitecture      string
	InstanceType         string    `gorm:"column:instance_type"`
	Storage              string
	Network              string
	OperatingSystem      string
	Type                 string
	Memory               string
	PhysicalProcessor    string    `gorm:"column:physical_processor"`
	MaxThroughput        string    `gorm:"column:max_throughput"`
	EnhancedNetworking   string    `gorm:"column:enhanced_networking"`
	GPU                  string    `gorm:"column:gpu"`
	MaxIOPS              string    `gorm:"column:max_iops"`
	CreatedDate          time.Time `gorm:"default:current_timestamp"`
	ModifiedDate         time.Time `gorm:"default:current_timestamp"`
	DisableFlag          bool      `gorm:"default:false"`
}

// ========== JSON API Models ==========

// For region API response
type APIRegion struct {
	Name string `json:"name"`
}

type RegionList struct {
	Items []APIRegion `json:"items"`
}

// Full response from the SKU API
type SkuResponse struct {
	Skus          []SkuItem `json:"skus"`
	NextPageToken string    `json:"nextPageToken"`
}

// Single SKU item
type SkuItem struct {
	SkuID               string         `json:"skuId"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	Category            SkuCategory    `json:"category"`
	ServiceRegions      []string       `json:"serviceRegions"`
	PricingInfo         []PricingInfo  `json:"pricingInfo"`
	ServiceProviderName string         `json:"serviceProviderName"`
	GeoTaxonomy         GeoTaxonomy    `json:"geoTaxonomy"`
}

// SKU Category section
type SkuCategory struct {
	ResourceFamily string `json:"resourceFamily"`
}

// Pricing section
type PricingInfo struct {
	PricingExpression PricingExpression `json:"pricingExpression"`
}

type PricingExpression struct {
	UsageUnit               string       `json:"usageUnit"`
	DisplayQuantity         int          `json:"displayQuantity"`
	TieredRates             []TieredRate `json:"tieredRates"`
	UsageUnitDescription    string       `json:"usageUnitDescription"`
	BaseUnit                string       `json:"baseUnit"`
	BaseUnitDescription     string       `json:"baseUnitDescription"`
	BaseUnitConversionFactor float64     `json:"baseUnitConversionFactor"`
}

type TieredRate struct {
	StartUsageAmount float64   `json:"startUsageAmount"`
	UnitPrice        UnitPrice `json:"unitPrice"`
}

type UnitPrice struct {
	CurrencyCode string `json:"currencyCode"`
	Units        string `json:"units"`
	Nanos        int    `json:"nanos"`
}

// Geo taxonomy for region mapping
type GeoTaxonomy struct {
	Type    string   `json:"type"`
	Regions []string `json:"regions"`
}
