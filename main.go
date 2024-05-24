package main

import (
	DB "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalServer "Phoenicia-Digital-Base-API/base/server"
)

func main() {
	if DB.Postgres.DB != nil {
		defer DB.Postgres.DB.Close()
	}
	PhoeniciaDigitalServer.StartServer()
}
