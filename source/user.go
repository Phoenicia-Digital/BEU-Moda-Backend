// source package `./souce/user.go` used for user & account functionality
package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UID         uint        `json:"ID"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	BillingInfo BillingInfo `json:"billing_info"`
	Session     Session     `json:"session_info"`
}

type Session struct {
	ID         uint      `json:"ID"`
	Session_id string    `json:"session_id"`
	Expires    time.Time `json:"expires"`
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

	loginUser.Email = r.URL.Query().Get("email")
	loginUser.Password = r.URL.Query().Get("password")
	if loginUser.Email == "" || loginUser.Password == "" {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Failed Dependancy Email: %s, Password: %s", loginUser.Email, loginUser.Password)}
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

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckExistingSession"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(loginUser.UID).Scan(&loginUser.Session.ID, &loginUser.Session.Session_id, &loginUser.Session.Expires); err != nil {
		if err == sql.ErrNoRows {

			loginUser.Session.Session_id = generateSessionID(loginUser.Email, loginUser.Password)
			loginUser.Session.Expires = time.Now().Add(15 * 24 * time.Hour).UTC()

			if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CreateNewSession"); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
			} else {
				if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
				} else {
					stmt = _stmt
					defer _stmt.Close()
				}
			}

			if err := stmt.QueryRow(loginUser.Session.Session_id, loginUser.UID, time.Now().UTC().Format(time.RFC3339), loginUser.Session.Expires).Scan(&loginUser.Session.ID); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:    "session_id",
					Value:   loginUser.Session.Session_id,
					Expires: loginUser.Session.Expires,
				})
				return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: loginUser}
			}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		}
	} else {

		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   loginUser.Session.Session_id,
			Expires: loginUser.Session.Expires,
		})
		return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: loginUser}

	}

}

func generateSessionID(email, password string) string {
	raw := fmt.Sprintf("%s:%s", email, password)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
