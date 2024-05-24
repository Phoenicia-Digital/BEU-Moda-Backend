// source package `./souce/user.go` used for user & account functionality
package source

type User struct {
	UID         uint        `json:"ID"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	BillingInfo BillingInfo `json:"billing info"`
}

func (u User) Account() (string, string) {
	return u.Email, u.Password
}
