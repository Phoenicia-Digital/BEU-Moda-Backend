// source package `./souce/catalogue.go` used to store catalogue item types & functionality
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

type Item struct {
	ID          uint    `json:"ID"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    uint16  `json:"quantity"`
	Color       string  `json:"image"`
}

func AddNewItem(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var newItem Item
	var admin User
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		admin.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			admin.UID = uint(uid)
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

	if err := stmt.QueryRow(admin.Session.Session_id, admin.UID).Scan(&admin.Session.ID); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NOT AN ADMIN"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Check Session | Error: %s", err.Error())}
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AddNewItem"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Retrieve session | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if newItem.Name == "" || newItem.Quantity <= 0.00 {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Name: %s OR Quantity: %d | NOT SUFFICIENT FOR ITEM TO BE ADDED", newItem.Name, newItem.Quantity)}
	}

	if err := stmt.QueryRow(newItem.Name, newItem.Description, newItem.Price, newItem.Quantity, newItem.Color).Scan(&newItem.ID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Add Item | Error: %s", err.Error())}
	}

	PhoeniciaDigitalUtils.Log(fmt.Sprintf("New Item Added To Database: %s, ID: %d | BY ADMIN ID: %d", newItem.Name, newItem.ID, admin.UID))
	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: newItem}

}

func EditItemByID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var editItem Item
	var admin User
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&editItem); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		editItem.ID = uint(val)
	}

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		admin.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			admin.UID = uint(uid)
		}
	}

	if editItem.ID <= 0 || editItem.Name == "" || editItem.Quantity == 0 || editItem.Price == 0 {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "Check Item ID, NAME, QUANTITY OR PRICE | VALUES NOT ACCEPTED"}
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

	if err := stmt.QueryRow(admin.Session.Session_id, admin.UID).Scan(&admin.Session.ID); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NOT AN ADMIN"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Check Session | Error: %s", err.Error())}
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("EditItem"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read SQL | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(editItem.Name, editItem.Description, editItem.Price, editItem.Quantity, editItem.Color, editItem.ID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Edit item | Error: %s", err.Error())}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Does NOT Exist", editItem.ID)}
			}
		}
	}

	PhoeniciaDigitalUtils.Log(fmt.Sprintf("Item With ID: %d EDITED, By Admin ID: %d", editItem.ID, admin.UID))
	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: editItem}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var deleteID uint
	var admin User
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		deleteID = uint(val)
	}

	if cooki, err := r.Cookie("session_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No Session ID | Error: %s", err.Error())}
	} else {
		admin.Session.Session_id = cooki.Value
	}

	if cookie, err := r.Cookie("user_id"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("No User ID | Error: %s", err.Error())}
	} else {
		if uid, err := strconv.Atoi(cookie.Value); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("User ID NOT an uint | Error: %s", err.Error())}
		} else {
			admin.UID = uint(uid)
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

	if err := stmt.QueryRow(admin.Session.Session_id, admin.UID).Scan(&admin.Session.ID); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusNotFound, Quote: "NOT AN ADMIN"}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Check Session | Error: %s", err.Error())}
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("DeleteItem"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read SQL | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(deleteID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Edit item | Error: %s", err.Error())}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Does NOT Exist", deleteID)}
			}
		}
	}

	PhoeniciaDigitalUtils.Log(fmt.Sprintf("Item With ID: %d DELETED, By Admin ID: %d", deleteID, admin.UID))
	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: fmt.Sprintf("Item With ID: %d DELETED", deleteID)}

}

func GetItemByID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var item Item
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		item.ID = uint(val)
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetItemByID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(item.ID).Scan(&item.Name, &item.Description, &item.Price, &item.Quantity, &item.Color); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item With ID: %d, DOES NOT EXIST", item.ID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Fetch Item ID: %d", item.ID)}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: item}
}

func GetItems(wh http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var items []Item
	var item Item
	var stmt *sql.Stmt

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetItems"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if rows, err := stmt.Query(); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed to get Items ROWS"}
	} else {
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Quantity, &item.Color); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Scanning Items | Error: %s", err.Error())}
			} else {
				items = append(items, item)
			}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: items}

}
