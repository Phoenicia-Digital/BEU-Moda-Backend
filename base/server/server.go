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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	})
	multiplexer.Handle("POST /register", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.RegisterNewUser))
	multiplexer.Handle("GET /register", PhoeniciaDigitalUtils.PhoeniciaDigitalHandler(source.LoginUser))
}
