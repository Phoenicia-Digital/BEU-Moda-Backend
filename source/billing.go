// source package `./souce/billing.go` used to store billing types & functionality
package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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
	ID          uint   `json:"ID"`
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

func ManageBillingInfo(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var newBillingInfo User
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&newBillingInfo.BillingInfo); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		newBillingInfo.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			newBillingInfo.UID = uint(uid)
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckSession"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(newBillingInfo.Session.Session_id, newBillingInfo.UID).Scan(&newBillingInfo.Session.ID); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NO SESSION"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Check Session | Error: %s", err.Error())}
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckBillingInfo"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(newBillingInfo.UID).Scan(&newBillingInfo.UID); err != nil {
		if err == sql.ErrNoRows {

			if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AddNewBillingInfo"); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
			} else {
				if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
				} else {
					stmt = _stmt
					defer _stmt.Close()
				}
			}

			if err := stmt.QueryRow(newBillingInfo.UID, newBillingInfo.BillingInfo.Country, newBillingInfo.BillingInfo.Province, newBillingInfo.BillingInfo.City,
				newBillingInfo.BillingInfo.Street, newBillingInfo.BillingInfo.Building, newBillingInfo.BillingInfo.Floor,
				newBillingInfo.BillingInfo.PhoneNumber, newBillingInfo.BillingInfo.FirstName, newBillingInfo.BillingInfo.LastName).Scan(&newBillingInfo.BillingInfo.ID); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
			}

			return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusCreated, Quote: newBillingInfo.BillingInfo}

		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("UpdateBillingInfo"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(newBillingInfo.BillingInfo.Country, newBillingInfo.BillingInfo.Province, newBillingInfo.BillingInfo.City,
		newBillingInfo.BillingInfo.Street, newBillingInfo.BillingInfo.Building, newBillingInfo.BillingInfo.Floor,
		newBillingInfo.BillingInfo.PhoneNumber, newBillingInfo.BillingInfo.FirstName, newBillingInfo.BillingInfo.LastName, newBillingInfo.UID).Scan(&newBillingInfo.BillingInfo.ID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: newBillingInfo.BillingInfo}
}

func FetchBillingInfo(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
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
			usr.UID = uint(uid)
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckSession"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
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

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("FetchBillingInfo"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(usr.UID).Scan(&usr.BillingInfo.ID, &usr.UID, &usr.BillingInfo.Country, &usr.BillingInfo.Province, &usr.BillingInfo.City, &usr.BillingInfo.Street,
		&usr.BillingInfo.Building, &usr.BillingInfo.Floor, &usr.BillingInfo.PhoneNumber, &usr.BillingInfo.FirstName, &usr.BillingInfo.LastName); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NO BILLING INFO"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Fetch Billing Info | Error: %s", err.Error())}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: usr.BillingInfo}

}
