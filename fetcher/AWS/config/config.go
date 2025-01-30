package config
 
// API Links
const (
    BaseURL         = "https://pricing.us-east-1.amazonaws.com"
    RegionURL       = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/region_index.json"
    SavingRegionURL = "https://pricing.us-east-1.amazonaws.com/savingsPlan/v1.0/aws/AWSComputeSavingsPlan/current/region_index.json"
    DbConnStr       = "host=localhost user=postgres password=password dbname=cloudcost sslmode=disable"
    PriceListPath   = "./price-list"
)