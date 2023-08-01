package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

var Collection *mongo.Collection

type User struct {
	Username   string   `bson:"username" json:"username"`
	ExpiryDate int64    `bson:"expiry_date" json:"expiry_date"`
	Outputs    []string `bson:"outputs" json:"outputs"`
	Password   string   `bson:"password" json:"-"`
}

func InitMongoDB() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongodbURL := os.Getenv("MONGODB_URL")

	clientOptions := options.Client().ApplyURI(mongodbURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %s", err)
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %s", err)
	}

	log.Println("Connected to DB")

	Collection = Client.Database("users").Collection("users")
}
