// File: `Server Handler Functions File` source/utils/handlers.go
package PhoeniciaDigitalUtils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// The Interface we use for our functions | Since they return an error
// used in http.Handle(_, <PhoeniciaDigitalHandler(function)>)
type PhoeniciaDigitalHandler func(http.ResponseWriter, *http.Request) PhoeniciaDigitalResponse

// Implement the `PhoeniciaDigitalResponse` Interface this will be the Responses on all API Functions Either
// Success Or Error and each will be handled automatically using the PhoeniciaDigitalHandler
type PhoeniciaDigitalResponse interface {
	Status() int
	Log() string
	Response() any
}

type ApiError struct {
	// ApiError struct can be returned as a error in API functions and automatically does error handling
	// for unsucessful calls || ERRORS via the ServeHTTP interface for `PhoeniciaDigitalHandler`
	Code  int `json:"status"`
	Quote any `json:"response"`
}

type ApiSuccess struct {
	// ApiSuccess struct can be returned as a `val` in `SendJSON` so we get a JSON response
	// for successful calls
	Code  int `json:"status"`
	Quote any `json:"response"`
}

// Implements the `phoeniciaDigitalResponse` interface for ApiError.
func (apiErr ApiError) Log() string {
	if str, ok := apiErr.Quote.(string); ok {
		// Handle string ErrorQuote
		return fmt.Sprintf("Error Code: %d, Error Quote: %s", apiErr.Code, str)
	} else {
		// Handle non-string ErrorQuote
		return fmt.Sprintf("Error Code: %d, Error Quote (NON STRING VALUE): %v", apiErr.Code, apiErr.Quote)
	}
}

func (apiErr ApiError) Status() int {
	return apiErr.Code
}

func (apiErr ApiError) Response() any {
	return apiErr.Quote
}

// Implements the `phoeniciaDigitalResponse` interface for ApiSuccess.
func (apiScc ApiSuccess) Log() string {
	if str, ok := apiScc.Quote.(string); ok {
		// Handle string ErrorQuote
		return fmt.Sprintf("Success Code: %d, Success Quote: %s", apiScc.Code, str)
	} else {
		// Handle non-string ErrorQuote
		return fmt.Sprintf("Success Code: %d, Success Quote (NON STRING VALUE): %v", apiScc.Code, apiScc.Quote)
	}
}

func (apiScc ApiSuccess) Status() int {
	return apiScc.Code
}

func (apiScc ApiSuccess) Response() any {
	return apiScc.Quote
}

// Implementing the SendJSON function that will send JSON responses in an organized consistant manner through
// the api responses
func SendJSON(w http.ResponseWriter, status int, val any) error {
	// Set the Response Writers Header status and the content type to JSON so that we can send JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow requests from any origin (http://localhost:3000 in your case)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(status)

	// Encode The Value `val` into the Response Writer and return an error if occured which will be managed by
	// the ServeHTTP interface ONLY IF USING `PhoeniciaDigitalHnadler`
	return json.NewEncoder(w).Encode(val)
}

// Implementaion of the http.ServeHTTP interface on `PhoeniciaDigitalHandler`
func (pdf PhoeniciaDigitalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow requests from any origin (http://localhost:3000 in your case)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Call the underlying handler function & Handle Responses | ERROR / NON-ERROR
	if response := pdf(w, r); response != nil {
		if _, converted := response.(ApiSuccess); converted {
			if response.Status() != 0 || response.Response() != nil {
				if ierr := SendJSON(w, response.Status(), response.Response()); ierr != nil {
					http.Error(w, response.Log(), response.Status())
				}
			} else {
				http.Error(w, "Request Returned Successful But Empty Paramaters", http.StatusInternalServerError)
			}
		} else if _, converted := response.(ApiError); converted {
			if response.Status() != 0 || response.Response() != nil {
				Log(response.Log())
				if ierr := SendJSON(w, response.Status(), response); ierr != nil {
					http.Error(w, response.Log(), response.Status())
				}
			} else {
				http.Error(w, "Request Returned Error But Empty Paramaters", http.StatusInternalServerError)
			}
		}
	} else {
		Log("!!!CAUTION!!! `PhoeniciaDigitalResponse` nil RETURNED ON LAST API CALL")
		if ierr := SendJSON(w, http.StatusInternalServerError, ApiError{Code: http.StatusInternalServerError, Quote: "!!!CAUTION!!! NO TYPE `PhoeniciaDigitalResponse` RETURNED ON LAST API CALL"}); ierr != nil {
			http.Error(w, "!!!CAUTION!!! `PhoeniciaDigitalResponse` nil RETURNED ON LAST API CALL", http.StatusInternalServerError)
		}
	}
}

//``` DISCONTINUED USED TO USE THIS AS THE HANDLER INSIDE THE http.HandleFunc but now we http.Handle with the
// PhoeniciaDigitalAPIFunction Interface! ```

// func Handler(df _PhoeniciaDigitalAPIFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 		df(w, r)
// 	}
// }
