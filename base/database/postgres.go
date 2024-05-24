package PhoeniciaDigitalDatabase

import (
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	PhoeniciaDigitalConfig "Phoenicia-Digital-Base-API/config"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type postgres struct {
	DB *sql.DB
}

var Postgres *postgres = &postgres{
	DB: implementPostgres(),
}

func (p postgres) ReadSQL(fileName string) (string, error) {
	if query, err := os.ReadFile(fmt.Sprintf("./sql/%s.sql", fileName)); err != nil {
		log.Printf("Error reading query file: %s | Error: %s", query, err.Error())
		PhoeniciaDigitalUtils.Log(fmt.Sprintf("Error reading query file: %s | Error: %s", query, err.Error()))
		return "", err
	} else {
		return string(query), err
	}
}

func implementPostgres() *sql.DB {
	if PhoeniciaDigitalConfig.Config.Postgres.Postgres_user != "" && PhoeniciaDigitalConfig.Config.Postgres.Postgres_db != "" {
		if PhoeniciaDigitalConfig.Config.Postgres.Postgres_password != "" {
			if PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl != "" {
				var conStr string = fmt.Sprintf("user=%s dbname=%s password=%s sslmode=%s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, PhoeniciaDigitalConfig.Config.Postgres.Postgres_password, PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl)
				if db, err := sql.Open("postgres", conStr); err != nil && db != nil {
					log.Fatalf("Failed to implement Postgres Database | Error: %s", err.Error())
					PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement Postgres Database | Error: %s", err.Error()))
					return nil
				} else {
					if err := db.Ping(); err != nil {
						log.Fatalf("Failed to connect to Postgres Database | Verify Postgres Database config values ./config/.env | Error: %s", err.Error())
						return nil
					} else {
						if rows, err := db.Query("SELECT 1"); err != nil {
							log.Fatalf("Database Name: %s Does NOT EXIST | Change at ./config/.env | Error: %s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, err.Error())
							return nil
						} else {
							rows.Close()
							log.Printf("Implemented Postgres Database connection settings: user=%s dbname=%s password=*** sslmode=%s\n", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl)
							return db
						}
					}
				}
			} else {
				var conStr string = fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, PhoeniciaDigitalConfig.Config.Postgres.Postgres_password)
				if db, err := sql.Open("postgres", conStr); err != nil && db != nil {
					log.Fatalf("Failed to implement Postgres Database | Error: %s", err.Error())
					PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement Postgres Database | Error: %s", err.Error()))
					return nil
				} else {
					if err := db.Ping(); err != nil {
						log.Fatalf("Failed to connect to Postgres Database | Verify Postgres Database config values ./config/.env | Error: %s", err.Error())
						return nil
					} else {
						if rows, err := db.Query("SELECT 1"); err != nil {
							log.Fatalf("Database Name: %s Does NOT EXIST | Change at ./config/.env | Error: %s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, err.Error())
							return nil
						} else {
							rows.Close()
							log.Printf("Implemented Postgres Database connection settings: user=%s dbname=%s password=*** sslmode=disable\n", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db)
							return db
						}
					}
				}
			}
		} else {
			if PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl != "" {
				var conStr string = fmt.Sprintf("user=%s dbname=%s sslmode=%s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl)
				if db, err := sql.Open("postgres", conStr); err != nil && db != nil {
					log.Fatalf("Failed to implement Postgres Database | Error: %s", err.Error())
					PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement Postgres Database | Error: %s", err.Error()))
					return nil
				} else {
					if err := db.Ping(); err != nil {
						log.Fatalf("Failed to connect to Postgres Database | Verify Postgres Database config values ./config/.env | Error: %s", err.Error())
						return nil
					} else {
						if rows, err := db.Query("SELECT 1"); err != nil {
							log.Fatalf("Database Name: %s Does NOT EXIST | Change at ./config/.env | Error: %s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, err.Error())
							return nil
						} else {
							rows.Close()
							log.Printf("Implemented Postgres Database connection settings: user=%s dbname=%s sslmode=%s\n", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, PhoeniciaDigitalConfig.Config.Postgres.Postgres_ssl)
							return db
						}
					}
				}
			} else {
				var conStr string = fmt.Sprintf("user=%s dbname=%s sslmode=disable", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db)
				if db, err := sql.Open("postgres", conStr); err != nil && db != nil {
					log.Fatalf("Failed to implement Postgres Database | Error: %s", err.Error())
					PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement Postgres Database | Error: %s", err.Error()))
					return nil
				} else {
					if err := db.Ping(); err != nil {
						log.Fatalf("Failed to connect to Postgres Database | Verify Postgres Database config values ./config/.env | Error: %s", err.Error())
						return nil
					} else {
						if rows, err := db.Query("SELECT 1"); err != nil {
							log.Fatalf("Database Name: %s Does NOT EXIST | Change at ./config/.env | Error: %s", PhoeniciaDigitalConfig.Config.Postgres.Postgres_db, err.Error())
							return nil
						} else {
							rows.Close()
							log.Printf("Implemented Postgres Database connection settings: user=%s dbname=%s sslmode=disable\n", PhoeniciaDigitalConfig.Config.Postgres.Postgres_user, PhoeniciaDigitalConfig.Config.Postgres.Postgres_db)
							return db
						}
					}
				}
			}
		}
	} else {
		log.Printf("Continued with No Postgres Database implementation! | In case expected a db connection POSTGRES_USER & POSTGRES_DB fields REQUIRED ./config/.env\n")
		PhoeniciaDigitalUtils.Log("Continued with No Postgres Database implementation! | In case expected a db connection POSTGRES_USER & POSTGRES_DB fields REQUIRED ./config/.env")
		return nil
	}
}
