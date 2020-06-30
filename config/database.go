package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kiranbhalerao123/gotter/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoInstance contains the mongo client and database object
type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

type MongoOptions struct {
	New options.FindOneAndUpdateOptions
}

var Mongo MongoInstance
var MongoOps MongoOptions

func init() {
	// Database settings
	dbName := utils.GoDotEnvVariable("DB_NAME")
	mongoURI := utils.GoDotEnvVariable("DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		log.Fatal(err)
	}

	// ping to db to check database connection is successfully established
	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		log.Fatal(err)
	}

	Mongo = MongoInstance{
		Client: client,
		DB:     db,
	}

	// setup some options var
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	MongoOps = MongoOptions{
		New: opt,
	}

	fmt.Println("Connected to MongoDB!")
}
