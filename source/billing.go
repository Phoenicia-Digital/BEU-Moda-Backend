// source package `./souce/billing.go` used to store billing types & functionality
package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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

	if err := json.NewDecoder(r.Body).Decode(&newBillingInfo); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckUserForBillingInfo"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(newBillingInfo.UID, newBillingInfo.Email).Scan(&newBillingInfo.UID, &newBillingInfo.Email); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNonAuthoritativeInfo, Quote: "NOT VALID USER"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
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

			return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusCreated, Quote: newBillingInfo}

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

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: newBillingInfo}
}
