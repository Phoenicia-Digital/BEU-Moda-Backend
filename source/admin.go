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
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterNewAdmin(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var newUser User
	var dbPassword string
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("RegisterNewAdmin"); err != nil {
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

	newUser.Session.Session_id = generateAdminSessionID(newUser.Email, newUser.Password)
	newUser.Session.Expires = time.Now().Add(15 * 24 * time.Hour).UTC()

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CreateNewAdminSession"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(newUser.Session.Session_id, newUser.UID, time.Now().UTC().Format(time.RFC3339), newUser.Session.Expires).Scan(&newUser.Session.ID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Create session | Error: %s", err.Error())}
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    newUser.Session.Session_id,
			Expires:  newUser.Session.Expires,
			Secure:   false,
			HttpOnly: true,
			Path:     "/",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_email",
			Value:    newUser.Email,
			Expires:  time.Now().AddDate(10, 0, 0),
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    fmt.Sprint(*newUser.UID),
			Expires:  time.Now().AddDate(10, 0, 0),
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		})
		return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: "Admin Created"}
	}
}

func LoginAdmin(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var loginUser User
	var dbPassword string
	var stmt *sql.Stmt

	loginUser.Email = r.URL.Query().Get("email")
	loginUser.Password = r.URL.Query().Get("password")
	if loginUser.Email == "" || loginUser.Password == "" {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Failed Dependancy Email: %s, Password: %s", loginUser.Email, loginUser.Password)}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("LoginAdmin"); err != nil {
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

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckExistingAdminSession"); err != nil {
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

			loginUser.Session.Session_id = generateAdminSessionID(loginUser.Email, loginUser.Password)
			loginUser.Session.Expires = time.Now().Add(15 * 24 * time.Hour).UTC()

			if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CreateNewAdminSession"); err != nil {
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
					Name:     "session_id",
					Value:    loginUser.Session.Session_id,
					Expires:  loginUser.Session.Expires,
					Secure:   false,
					HttpOnly: true,
					Path:     "/",
				})

				http.SetCookie(w, &http.Cookie{
					Name:     "user_email",
					Value:    loginUser.Email,
					Expires:  time.Now().AddDate(10, 0, 0),
					Secure:   false,
					HttpOnly: false,
					Path:     "/",
				})

				http.SetCookie(w, &http.Cookie{
					Name:     "user_id",
					Value:    fmt.Sprint(*loginUser.UID),
					Expires:  time.Now().AddDate(10, 0, 0),
					Secure:   false,
					HttpOnly: false,
					Path:     "/",
				})
				return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: "Session Created"}
			}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		}
	} else {

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    loginUser.Session.Session_id,
			Expires:  loginUser.Session.Expires,
			Secure:   false,
			HttpOnly: true,
			Path:     "/",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_email",
			Value:    loginUser.Email,
			Expires:  time.Now().AddDate(10, 0, 0),
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    fmt.Sprint(*loginUser.UID),
			Expires:  time.Now().AddDate(10, 0, 0),
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		})

		return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: "Session Exists"}

	}

}

func CheckAdminSession(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var stmt *sql.Stmt

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		usr.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			usr.UID = new(uint)
			*usr.UID = uint(uid)
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckAdminSession"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(usr.Session.Session_id, usr.UID).Scan(&usr.Session.ID); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NO SESSION"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Check Session | Error: %s", err.Error())}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: "Session Exists"}

}

func LogoutAdmin(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var stmt *sql.Stmt

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		usr.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			usr.UID = new(uint)
			*usr.UID = uint(uid)
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("LogoutAdmin"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(usr.Session.Session_id, usr.UID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Delete Session | Error: %s", err.Error())}
	} else {
		if rows, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else if rows == 0 {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NO SESSION"}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour).UTC(),
		Secure:   false,
		HttpOnly: true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_email",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour).UTC(),
		Secure:   false,
		HttpOnly: false,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour).UTC(),
		Secure:   false,
		HttpOnly: false,
		Path:     "/",
	})

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: "Session Deleted"}

}

func generateAdminSessionID(email, password string) string {
	raw := fmt.Sprintf("%s:%s:BeuModaAdmin", email, password)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
