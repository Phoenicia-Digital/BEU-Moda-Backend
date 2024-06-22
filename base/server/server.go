// File: `Server Implementation File` source/server/server.go
package PhoeniciaDigitalServer

import (
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	PhoeniciaDigitalConfig "Phoenicia-Digital-Base-API/config"
	"Phoenicia-Digital-Base-API/source"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Initialize Server Ecosystem Variables
var multiplexer *http.ServeMux = http.NewServeMux()

var PhoeniciaDigitalServer *http.Server = &http.Server{
	Addr:    PhoeniciaDigitalConfig.Config.Port,
	Handler: multiplexer,
}

func StartServer() {
	if PhoeniciaDigitalServer.Addr != ":" {
		if portNumber, err := strconv.Atoi(PhoeniciaDigitalServer.Addr[1:]); err != nil {
			log.Printf("Given PORT is Invalid: %s != int | Change in ./config/.env", PhoeniciaDigitalServer.Addr[1:])
			PhoeniciaDigitalUtils.Log(fmt.Sprintf("Given PORT is Invalid: %s != int | Change in ./config/.env", PhoeniciaDigitalServer.Addr[1:]))
		} else {
			if portNumber >= 0 && portNumber <= 65535 {
				log.Printf("Server Running on http://localhost%s", PhoeniciaDigitalServer.Addr)
				PhoeniciaDigitalUtils.Log(fmt.Sprintf("Server started on PORT --> %s", PhoeniciaDigitalServer.Addr))
				log.Fatal(PhoeniciaDigitalServer.ListenAndServe())
			} else {
				log.Printf("Given PORT: %s is OUT OF RANGE 0 --> 65535 | Change in ./config/.env", PhoeniciaDigitalServer.Addr[1:])
				PhoeniciaDigitalUtils.Log(fmt.Sprintf("Given PORT: %s is OUT OF RANGE 0 --> 65535 | Change in ./config/.env", PhoeniciaDigitalServer.Addr[1:]))
			}
		}
	} else {
		log.Printf("Given PORT is empty | Change in ./config/.env")
		PhoeniciaDigitalUtils.Log("Given PORT is empty | Change in ~/config/.env")
	}
}

// Initialize Server Logic
func init() {
	multiplexer.HandleFunc("OPTIONS /register", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("POST /register", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.RegisterNewUser))

	multiplexer.HandleFunc("OPTIONS /login", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.CheckSession))
	multiplexer.Handle("POST /login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.LoginUser))
	multiplexer.Handle("DELETE /login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.LogoutUser))

	multiplexer.HandleFunc("OPTIONS /admin/register", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("POST /admin/register", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.RegisterNewAdmin))

	multiplexer.HandleFunc("OPTIONS /admin/login", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /admin/login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.CheckAdminSession))
	multiplexer.Handle("POST /admin/login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.LoginAdmin))
	multiplexer.Handle("DELETE /admin/login", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.LogoutAdmin))

	multiplexer.HandleFunc("OPTIONS /billing", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /billing", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.FetchBillingInfo))
	multiplexer.Handle("POST /billing", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.ManageBillingInfo))

	multiplexer.HandleFunc("OPTIONS /catalogue", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /catalogue", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.GetItems))
	multiplexer.Handle("POST /catalogue", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.AddNewItem))

	multiplexer.HandleFunc("OPTIONS /catalogue/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /catalogue/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.GetItemByID))
	multiplexer.Handle("PUT /catalogue/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.EditItemByID))
	multiplexer.Handle("DELETE /catalogue/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.DeleteItem))

	multiplexer.HandleFunc("OPTIONS /order", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /order", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.GetPendingOrdersByUserID))
	multiplexer.Handle("POST /order", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.PorcessOrder))

	multiplexer.HandleFunc("OPTIONS /order/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /order/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.GetPendingOrderByOrderID))

	multiplexer.HandleFunc("OPTIONS /admin/order", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /admin/order", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.GetPendingOrders))

	multiplexer.HandleFunc("OPTIONS /admin/order/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("GET /admin/order/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.AdminGetPendingOrderByOrderID))
	multiplexer.Handle("DELETE /admin/order/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.RemovePendingOrderByID))

	multiplexer.HandleFunc("OPTIONS /admin/order/complete/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests (can be more specific if needed)
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin (http://localhost:3000 in your case)
		w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("PUT /admin/order/complete/{id}", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.CompletePendingOrderByID))
}
