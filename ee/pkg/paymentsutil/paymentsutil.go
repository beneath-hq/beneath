package paymentsutil

import "time"

// IsBlacklisted checks to see if a country is sanctioned by the EU
func IsBlacklisted(country string) bool {
	blacklist := []string{"North Korea"} // +Iran, +Syria, +Russia, +Ukraine? see: https://sanctionsmap.eu/#/main, need to read details
	for _, sanctioned := range blacklist {
		if country == sanctioned {
			return true
		}
	}
	return false
}

// ComputeTaxPercentage computes the tax percent out of 100 to be added to each invoice item
func ComputeTaxPercentage(country string, region string, isCompany bool, date time.Time) (int, string) {
	if country == "Denmark" {
		return 25, "VAT Denmark"
	}
	return 0, "N/A"
}
