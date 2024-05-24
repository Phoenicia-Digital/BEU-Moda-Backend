// source package `./souce/catalogue.go` used to store catalogue item types & functionality
package source

type Item struct {
	UID         uint    `json:"ID"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    uint16  `json:"quantity"`
	Color       string  `json:"color"`
}
