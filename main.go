package main

import (
	PhoeniciaDigitalDatabase "Phoenicia-Digital-Base-API/base/database"
	PhoeniciaDigitalServer "Phoenicia-Digital-Base-API/base/server"
)

func main() {
	defer PhoeniciaDigitalDatabase.Postgres.DB.Close()
	PhoeniciaDigitalServer.StartServer()
}
