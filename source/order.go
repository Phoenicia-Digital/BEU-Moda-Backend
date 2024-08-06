package source

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type OrderedItem struct {
	ID       uint   `json:"ID"`
	Quantity uint16 `json:"quantity"`
}

type Order struct {
	OrderID      uint          `json:"order_id"`
	UserID       uint          `json:"user_id"`
	OrderedItems []OrderedItem `json:"ordered_items"`
	TotalPrice   float64       `json:"total_price"`
	OrderTime    time.Time     `json:"order_time"`
}

// Items added to a new table pending when order recieved Items go to history table
// When items added to new table pending you get items with the id added to order
// Change Quantity from items table check it quantity == 0 remove the row
// If ORDER canceled items go back to original quantity and delete from row

func PorcessOrder(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
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

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if len := len(order.OrderedItems); len <= 0 {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "Order Is Empty"}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckSession"); err != nil {
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
	} else {
		order.UserID = usr.UID
	}

	// Start the validation of the Order to make sure no injections or errors occure
	// and add the data to order

	for _, orderedItem := range order.OrderedItems {
		var price float64
		var availableQuantity uint16

		if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetItemByIDForOrder"); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
		} else {
			if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
			} else {
				stmt = _stmt
				defer _stmt.Close()
			}
		}

		if err := stmt.QueryRow(orderedItem.ID).Scan(&price, &availableQuantity); err != nil {
			if err == sql.ErrNoRows {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item With ID: %d, DOES NOT EXIST", orderedItem.ID)}
			} else {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Fetch Item ID: %d", orderedItem.ID)}
			}
		}

		if orderedItem.Quantity > availableQuantity {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("Item With ID: %d | Quantity Ordered More than Available", orderedItem.ID)}
		} else {
			order.TotalPrice += price * float64(orderedItem.Quantity)
		}
	}

	// Start the Update of the Items Ordered | Changing the new quantity

	for _, orderedItem := range order.OrderedItems {

		if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("UpdateItemAfterOrder"); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
		} else {
			if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
			} else {
				stmt = _stmt
				defer _stmt.Close()
			}
		}

		if res, err := stmt.Exec(orderedItem.Quantity, orderedItem.ID); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Update Item ID: %d", orderedItem.ID)}
		} else {
			if rowsAffected, err := res.RowsAffected(); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
			} else {
				if rowsAffected == 0 {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Failed To Update", orderedItem.ID)}
				}
			}
		}
	}

	// Finally add the new order to the database after the validation has been complete

	if orderedItemsJSON, err := json.Marshal(order.OrderedItems); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Marshal Ordered Items to JSON"}
	} else {
		jsonbItems = orderedItemsJSON
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AddNewOrder"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(order.UserID, jsonbItems, order.TotalPrice, time.Now()).Scan(&order.OrderID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Add New Order To Database"}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: fmt.Sprintf("New Order added with ID: %d", order.OrderID)}
}

func GetPendingOrdersByUserID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var orders []Order
	var order Order
	var jsonbItems []byte
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

	// Check User Session

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckSession"); err != nil {
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

	// Retrieve Order

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetPendingOrdersByUserID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if rows, err := stmt.Query(usr.UID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Querying Orders | Error: %s", err.Error())}
	} else {
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&order.OrderID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
				if err == sql.ErrNoRows {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: "NO Orders For User"}
				} else {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Fetch Order For User With ID: %d", usr.UID)}
				}
			} else {
				if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
				}
				orders = append(orders, order)
			}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: orders}
}

func GetPendingOrderByOrderID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		order.OrderID = uint(val)
	}

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

	// Check User Session

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CheckSession"); err != nil {
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

	// Retrieve Order

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetPendingOrderByOrderID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(usr.UID, order.OrderID).Scan(&jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("No Order With ID: %d For this USER", order.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Quering Row | Error: %s", err.Error())}
		}
	} else {
		if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: order}
}

func GetPendingOrders(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var orders []Order
	var order Order
	var jsonbItems []byte
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

	// Check Admin Session

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

	// Fetch All Items

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetPendingOrders"); err != nil {
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
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Query Rows"}
	} else {
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&order.OrderID, &order.UserID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Query Row | Error: %s", err.Error())}
			}

			if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
			}

			orders = append(orders, order)
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: orders}
}

func AdminGetPendingOrderByOrderID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		order.OrderID = uint(val)
	}

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

	// Check User Session

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

	// Retrieve Order

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AdminGetPendingOrderByOrderID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(order.OrderID).Scan(&order.UserID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("No Order With ID: %d", order.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Quering Row | Error: %s", err.Error())}
		}
	} else {
		if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: order}
}

func RemovePendingOrderByID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		order.OrderID = uint(val)
	}

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

	// Check Admin Session

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

	// GET THE EXISTING ORDER FIRST

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AdminGetPendingOrderByOrderID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(order.OrderID).Scan(&order.UserID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("No Order With ID: %d", order.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Quering Row | Error: %s", err.Error())}
		}
	} else {
		if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
		}
	}

	// Start the validation Process OF Edit

	for _, item := range order.OrderedItems {
		if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("UpdateItemAfterDeletedOrder"); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
		} else {
			if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
			} else {
				stmt = _stmt
				defer _stmt.Close()
			}
		}

		if res, err := stmt.Exec(item.Quantity, item.ID); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Update Item ID: %d", item.ID)}
		} else {
			if rowsAffected, err := res.RowsAffected(); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
			} else {
				if rowsAffected == 0 {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Failed To Update", item.ID)}
				}
			}
		}
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("RemoveOrderByID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(order.OrderID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Delete Order ID: %d", order.OrderID)}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Order with ID: %d Failed To Delete", order.OrderID)}
			}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: fmt.Sprintf("Order With ID: %d Deleted", order.OrderID)}

}

func CompletePendingOrderByID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		order.OrderID = uint(val)
	}

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

	// Check Admin Session

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

	// Get Order Details

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("AdminGetPendingOrderByOrderID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(order.OrderID).Scan(&order.UserID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("No Order With ID: %d", order.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Quering Row | Error: %s", err.Error())}
		}
	}

	// Add Order Details To History Database

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("CompletePendingOrderByID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(order.OrderID, order.UserID, jsonbItems, order.TotalPrice, time.Now()); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Add Order ID: %d To History", order.OrderID)}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Order with ID: %d Failed To Add To History", order.OrderID)}
			}
		}
	}

	// Delete Order From Pending Orders

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("RemoveOrderByID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(order.OrderID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Delete Order ID: %d", order.OrderID)}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Order with ID: %d Failed To Delete", order.OrderID)}
			}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: fmt.Sprintf("Order With ID: %d Completed", order.OrderID)}
}

func GetCompletedOrders(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var orders []Order
	var limit int
	var offset int
	var stmt *sql.Stmt

	if lim, err := strconv.Atoi(r.URL.Query().Get("limit")); err != nil {
		if r.URL.Query().Get("limit") == "" {
			limit = 20
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Provided Limit: %s NOT AN INT", r.URL.Query().Get("limit"))}
		}
	} else {
		limit = lim
	}

	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err != nil {
		if r.URL.Query().Get("offset") == "" {
			offset = 0
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Provided Offset: %s NOT AN INT", r.URL.Query().Get("offset"))}
		}
	} else {
		offset = off
	}

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

	// Check Admin Session

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

	// Get Orders By Limit & Offset

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetCompletedOrder"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read QUERY | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare QUERY | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if rows, err := stmt.Query(limit, offset); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed Query ROWS | Error: %s", err.Error())}
	} else {
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&order.OrderID, &order.UserID, &order.TotalPrice, &order.OrderTime); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed to Fetch Order From Rows"}
			}

			orders = append(orders, order)
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: orders}
}

func GetCompletedOrderByID(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var order Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		order.OrderID = uint(val)
	}

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

	// Check Admin Session

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

	// Get Order By Order ID

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetCompletedOrderByID"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read QUERY | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare QUERY | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(order.OrderID).Scan(&order.UserID, &jsonbItems, &order.TotalPrice, &order.OrderTime); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Order ID: %d Does NOT EXIST", order.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Faield To Query Row"}
		}
	}

	if err := json.Unmarshal(jsonbItems, &order.OrderedItems); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusOK, Quote: order}
}

func EditPendingOrderByAdmin(w http.ResponseWriter, r *http.Request) PhoeniciaDigitalUtils.PhoeniciaDigitalResponse {
	var usr User
	var originalOrder Order
	var originalQuant map[uint]uint16
	var differance map[uint]int16
	var editedOrder Order
	var jsonbItems []byte
	var stmt *sql.Stmt

	if err := json.NewDecoder(r.Body).Decode(&editedOrder); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: err.Error()}
	}

	if val, err := strconv.ParseUint(r.PathValue("id"), 10, 16); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("ID: %s NOT AN ID", r.PathValue("id"))}
	} else {
		originalOrder.OrderID, editedOrder.OrderID = uint(val), uint(val)
	}

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

	// Check Admin Session

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

	// Retrieve Original Order

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("EditPendingOrderFetch"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if err := stmt.QueryRow(originalOrder.OrderID).Scan(&jsonbItems, &originalOrder.TotalPrice); err != nil {
		if err == sql.ErrNoRows {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("No Order With ID: %d For this USER", originalOrder.OrderID)}
		} else {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Error Quering Row | Error: %s", err.Error())}
		}
	} else {
		if err := json.Unmarshal(jsonbItems, &originalOrder.OrderedItems); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Unmarshal Ordered Items"}
		}
	}

	originalQuant = originalQuantities(&originalOrder)

	// Start the validation of the Order to make sure no injections or errors occure
	// and add the data to order

	for _, orderedItem := range editedOrder.OrderedItems {
		var price float64
		var availableQuantity uint16

		if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("GetItemByIDForOrder"); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
		} else {
			if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
			} else {
				stmt = _stmt
				defer _stmt.Close()
			}
		}

		if err := stmt.QueryRow(orderedItem.ID).Scan(&price, &availableQuantity); err != nil {
			if err == sql.ErrNoRows {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item With ID: %d, DOES NOT EXIST", orderedItem.ID)}
			} else {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Fetch Item ID: %d", orderedItem.ID)}
			}
		}

		if orderedItem.Quantity > availableQuantity+originalQuant[orderedItem.ID] {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusConflict, Quote: fmt.Sprintf("Item With ID: %d | Quantity Ordered More than Available", orderedItem.ID)}
		} else {
			editedOrder.TotalPrice += price * float64(orderedItem.Quantity)
		}
	}

	differance = compareDifferance(&originalOrder, &editedOrder)

	for id, quantity := range differance {

		if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("UpdateItemAfterOrder"); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read Query | Error: %s", err.Error())}
		} else {
			if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
			} else {
				stmt = _stmt
				defer _stmt.Close()
			}
		}

		if res, err := stmt.Exec(quantity, id); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed To Update Item ID: %d", id)}
		} else {
			if rowsAffected, err := res.RowsAffected(); err != nil {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
			} else {
				if rowsAffected == 0 {
					return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Failed To Update", id)}
				}
			}
		}

	}

	if orderedItemsJSON, err := json.Marshal(editedOrder.OrderedItems); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: "Failed To Marshal Ordered Items to JSON"}
	} else {
		jsonbItems = orderedItemsJSON
	}

	if query, err := PhoeniciaDigitalDatabase.Postgres.ReadSQL("EditPendingOrder"); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Read SQL | Error: %s", err.Error())}
	} else {
		if _stmt, err := PhoeniciaDigitalDatabase.Postgres.DB.Prepare(query); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Unable to Prepare Query | Error: %s", err.Error())}
		} else {
			stmt = _stmt
			defer _stmt.Close()
		}
	}

	if res, err := stmt.Exec(jsonbItems, editedOrder.TotalPrice, originalOrder.OrderID); err != nil {
		return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Edit item | Error: %s", err.Error())}
	} else {
		if rowsAffected, err := res.RowsAffected(); err != nil {
			return PhoeniciaDigitalUtils.ApiError{Code: http.StatusInternalServerError, Quote: fmt.Sprintf("Failed to Check Affected Rows | Error: %s", err.Error())}
		} else {
			if rowsAffected == 0 {
				return PhoeniciaDigitalUtils.ApiError{Code: http.StatusFailedDependency, Quote: fmt.Sprintf("Item with ID: %d Does NOT Exist", originalOrder.OrderID)}
			}
		}
	}

	return PhoeniciaDigitalUtils.ApiSuccess{Code: http.StatusAccepted, Quote: "Pending Order Edited"}

}

// IN APP FUNCTIONS

func compareDifferance(originalOrder, editedOrder *Order) map[uint]int16 {
	var differance map[uint]int16 = map[uint]int16{}

	for _, item := range editedOrder.OrderedItems {
		differance[item.ID] = int16(item.Quantity)
	}

	for _, item := range originalOrder.OrderedItems {
		differance[item.ID] -= int16(item.Quantity)
	}

	return differance
}

func originalQuantities(originalOrder *Order) map[uint]uint16 {
	var originalQuantities map[uint]uint16 = map[uint]uint16{}

	for _, item := range originalOrder.OrderedItems {
		originalQuantities[item.ID] = uint16(item.Quantity)
	}

	return originalQuantities
}
