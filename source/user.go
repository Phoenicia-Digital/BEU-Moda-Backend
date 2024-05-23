// source package `./souce/user.go` used for user & account functionality
package source

type User struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Billing  BillingInfo `json:"billing info"`
}
