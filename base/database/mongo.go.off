package PhoeniciaDigitalDatabase

import (
	PhoeniciaDigitalUtils "Phoenicia-Digital-Base-API/base/utils"
	PhoeniciaDigitalConfig "Phoenicia-Digital-Base-API/config"
	"context"
	"fmt"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongodb struct {
	Client     *mongo.Client
	DB         *mongo.Database
	Collection *mongo.Collection
}

var Mongo *mongodb = implementMongo()

func implementMongo() *mongodb {
	if PhoeniciaDigitalConfig.Config.Mongo.Mongo_host != "" && PhoeniciaDigitalConfig.Config.Mongo.Mongo_port != "" {
		if portNum, err := strconv.Atoi(PhoeniciaDigitalConfig.Config.Mongo.Mongo_port); err != nil {
			PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement MongoDB client | Invalid Port Number: %s", PhoeniciaDigitalConfig.Config.Mongo.Mongo_port))
			log.Fatalf("Failed to implement MongoDB client | Invalid Port Number: %s", PhoeniciaDigitalConfig.Config.Mongo.Mongo_port)
			return nil
		} else {
			if portNum >= 0 && portNum <= 65535 {
				var clientStr string = fmt.Sprintf("mongodb://%s:%s", PhoeniciaDigitalConfig.Config.Mongo.Mongo_host, PhoeniciaDigitalConfig.Config.Mongo.Mongo_port)
				if clientConnection, err := mongo.Connect(context.Background(), options.Client().ApplyURI(clientStr)); err != nil {
					PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to create MongoDB client | Verify MONGODB_HOST: %s | Verify MONGODB_PORT: %s", PhoeniciaDigitalConfig.Config.Mongo.Mongo_host, PhoeniciaDigitalConfig.Config.Mongo.Mongo_port))
					log.Fatalf("Failed to create MongoDB client | Verify MONGODB_HOST: %s | Verify MONGODB_PORT: %s", PhoeniciaDigitalConfig.Config.Mongo.Mongo_host, PhoeniciaDigitalConfig.Config.Mongo.Mongo_port)
					return nil
				} else {
					if err := clientConnection.Ping(context.Background(), nil); err != nil {
						PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to connect MongoDB client: %s | Service might be down or WRONG PORT | ERROR: %s", clientStr, err.Error()))
						log.Fatalf("Failed to connect MongoDB client: %s | Service might be down or WRONG PORT | ERROR: %s", clientStr, err.Error())
						return nil
					} else {
						if PhoeniciaDigitalConfig.Config.Mongo.Mongo_db != "" {
							clientDatabase := clientConnection.Database(PhoeniciaDigitalConfig.Config.Mongo.Mongo_db)
							if err := clientDatabase.Client().Ping(context.Background(), nil); err != nil {
								PhoeniciaDigitalUtils.Log(fmt.Sprintf("Connected to MongoDB client but Database: %s Does NOT EXIST | Verify MONGODB_DATABASE", PhoeniciaDigitalConfig.Config.Mongo.Mongo_db))
								log.Fatalf("Connected to MongoDB client but Database: %s Does NOT EXIST | Verify MONGODB_DATABASE", PhoeniciaDigitalConfig.Config.Mongo.Mongo_db)
								return nil
							} else {
								if PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection != "" {
									if clientCollections, err := clientDatabase.ListCollectionNames(context.Background(), bson.D{{}}); err != nil {
										PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to list collection names to verify collection existance | ERROR: %s", err.Error()))
										log.Fatalf("Failed to list collection names to verify collection existance | ERROR: %s", err.Error())
										return nil
									} else {
										for _, collection := range clientCollections {
											if collection == PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection {
												clientCollection := clientDatabase.Collection(PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection)
												log.Printf("Implemented MongoDB Connection Client: %s, Database: %s, Collection: %s", clientStr, PhoeniciaDigitalConfig.Config.Mongo.Mongo_db, PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection)
												return &mongodb{
													Client:     clientConnection,
													DB:         clientDatabase,
													Collection: clientCollection,
												}
											}
										}
										PhoeniciaDigitalUtils.Log(fmt.Sprintf("Connect to MongoDB client & database but Collection: %s Does NOT EXIST | Verify MONGODB_COLLECTION", PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection))
										log.Fatalf("Connect to MongoDB client & database but Collection: %s Does NOT EXIST | Verify MONGODB_COLLECTION", PhoeniciaDigitalConfig.Config.Mongo.Mongo_collection)
										return nil
									}
								} else {
									PhoeniciaDigitalUtils.Log("Connected to MongoDB client & Database but NO Collection Provided | Verify MONGODB_COLLECTION")
									log.Fatal("Connected to MongoDB client & Database but NO Collection Provided | Verify MONGODB_COLLECTION")
									return nil
								}
							}
						} else {
							PhoeniciaDigitalUtils.Log("Connected to MongoDB client but NO Database Provided | Verify MONGODB_DATABASE")
							log.Fatal("Connected to MongoDB client but NO Database Provided | Verify MONGODB_DATABASE")
							return nil
						}
					}
				}
			} else {
				PhoeniciaDigitalUtils.Log(fmt.Sprintf("Failed to implement MongoDB client | Port: %d out of range 0 - 65535", portNum))
				log.Fatalf("Failed to implement MongoDB client | Port: %d out of range 0 - 65535", portNum)
				return nil
			}
		}
	} else {
		PhoeniciaDigitalUtils.Log("Continued with No MongoDB Database implementation! | In case expected a db connection MONGODB_HOST & MONGODB_PORT fields REQUIRED ./config/.env")
		log.Printf("Continued with No MongoDB Database implementation! | In case expected a db connection MONGODB_HOST & MONGODB_PORT fields REQUIRED ./config/.env\n")
		return nil
	}
}

func init() {
	if Mongo != nil && Mongo.Client != nil && Mongo.DB != nil && Mongo.Collection != nil {
		defer Mongo.Client.Disconnect(context.Background())
	}
}
