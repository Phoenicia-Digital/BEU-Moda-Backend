// source package `./souce/user.go` used for user & account functionality
package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"database/sql"
	"encoding/json"
	"fmt"
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
	var dbPassword string
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("RegisterNewUser"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if !strings.Contains(newUser.Email, "@gmail.com") {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "NOT AN EMAIL"}
	}

	if hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		dbPassword = string(hashed)
	}

	if err := stmt.QueryRow(newUser.Email, dbPassword).Scan(&newUser.UID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusCreated, Quote: newUser}
}

func LoginUser(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var loginUser User
	var dbPassword string
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&loginUser); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("LoginUser"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if !strings.Contains(loginUser.Email, "@gmail.com") {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "NOT AN EMAIL"}
	}

	if err := stmt.QueryRow(loginUser.Email).Scan(&loginUser.UID, &dbPassword); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: fmt.Sprintf("Email: %s Does Not Exist", loginUser.Email)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(loginUser.Password)); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusUnauthorized, Quote: "Invalid Password"}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusFound, Quote: loginUser}
}
