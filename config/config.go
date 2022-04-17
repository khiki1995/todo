package config

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"

	"log"
)

const (
	ListTableName     = "list"
	ListItemTableName = "item"
)

func Config(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Cannot load given key: " + key)
	}

	return os.Getenv(key)
}

type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

var MI MongoInstance

type Collection struct {
	ToDoList     *mongo.Collection
	ToDoListItem *mongo.Collection
}

var Collections Collection

func ConnectDB() {
	dbHost := Config("DB_HOST")
	dbPort := Config("DB_PORT")
	dbName := Config("DB_NAME")

	connectionString := fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort)
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.Database(dbName)

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	Collections = Collection{
		ToDoList:     db.Collection(ListTableName),
		ToDoListItem: db.Collection(ListItemTableName),
	}

	MI = MongoInstance{
		Client: client,
		DB:     db,
	}
}