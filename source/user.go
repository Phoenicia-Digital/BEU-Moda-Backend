// source package `./souce/user.go` used for user & account functionality
package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UID         uint        `json:"ID"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	BillingInfo BillingInfo `json:"billing info"`
}

func (u User) Account() (string, string) {
	return u.Email, u.Password
}

func RegisterNewUser(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}
	if !strings.Contains(newUser.Email, "@") || !strings.Contains(newUser.Email, ".") {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "NOT AN EMAIL"}
	}
	if hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		newUser.Password = string(hashed)
	}
	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("RegisterNewUser"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if err := PhoeniciaDigitalDatabase.Postgres.DB.QueryRow(query, newUser.Email, newUser.Password).Scan(&newUser.UID); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		}
	}
	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusCreated, Quote: newUser}
}
