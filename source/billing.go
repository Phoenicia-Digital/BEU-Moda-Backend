// source package `./souce/billing.go` used to store billing types & functionality
package source

import "fmt"

type BillingInfo struct {
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Street      string `json:"street"`
	Building    string `json:"building"`
	Floor       string `json:"floor"`
	PhoneNumber uint32 `json:"phone number"`
	FirstName   string `json:"first name"`
	LastName    string `json:"last name"`
}

func (b BillingInfo) Contact() uint32 {
	return b.PhoneNumber
}

func (b BillingInfo) Name() string {
	return fmt.Sprintf("%s %s", b.FirstName, b.LastName)
}

func (b BillingInfo) Address() string {
	if b.Country == "Lebanon" {
		return fmt.Sprintf("%s, %s St. %s Bld. %s floor.", b.City, b.Street, b.Building, b.Floor)
	} else {
		return fmt.Sprintf("%s %s %s, %s St. %s Bld. %s floor.", b.Country, b.Province, b.City, b.Street, b.Building, b.Floor)
	}
}
