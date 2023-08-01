package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define a MongoDB client variable
var client *mongo.Client

// Define a MongoDB collection variable
var collection *mongo.Collection

func initMongoDB() {
	// Set up MongoDB connection options
	clientOptions := options.Client().ApplyURI("mongodb+srv://ahmed2:Ghtwhts786@webtest.dipaskp.mongodb.net/?retryWrites=true&w=majority")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %s", err)
	}

	// Check if the connection was successful
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %s", err)
	}

	log.Println("Connected to DB")
	// Assign the MongoDB collection to the 'collection' variable
	collection = client.Database("users").Collection("users")
}

// Define a struct to hold the data sent in the request body
type User struct {
	Username   string   `bson:"username" json:"username"`
	ExpiryDate int64    `bson:"expiry_date" json:"expiry_date"`
	Outputs    []string `bson:"outputs" json:"outputs"`
	Password   string   `bson:"password" json:"-"`
}

// ...

func addUser(c *fiber.Ctx) error {
	// Create a new User instance to hold the request body data
	user := new(User)

	// Parse the request body into the User struct
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request format",
		})
	}

	log.Println(user.ExpiryDate)
	// Check if the expiry date is valid
	// expiryTime := time.Unix(user.Expiry, 0)
	// if expiryTime.Before(time.Now()) {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"message": "Expiry date should be in the future",
	// 	})
	// }

	// Define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the new user into the MongoDB collection
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error adding user",
		})
	}

	// Send the inserted user ID as the response
	return c.Status(201).JSON(user)
}

func getUserByUsername(c *fiber.Ctx) error {
	// Get the username from the path parameters
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Username not provided",
		})
	}

	// Define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a filter to find the user by the given username
	filter := bson.M{"username": username}

	// Query the MongoDB collection
	var user User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  false,
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error retrieving user",
		})
	}

	return c.JSON(user)
}

type UserPatch struct {
	Username   string   `json:"username,omitempty"`
	ExpiryDate int64    `json:"expiry_date,omitempty"`
	Outputs    []string `json:"outputs,omitempty"`
	Password   string   `json:"password,omitempty"`
}

func updateUserByUsername(c *fiber.Ctx) error {
	// Get the username from the path parameters
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Username not provided",
		})
	}

	// Create a new UserPatch instance to hold the request body data
	patchData := new(UserPatch)

	// Parse the request body into the UserPatch struct
	if err := c.BodyParser(patchData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request format",
		})
	}

	// Define a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a filter to find the user by the given username
	filter := bson.M{"username": username}

	// Create an update with the fields to be modified
	update := bson.M{}
	if patchData.Username != "" {
		update["username"] = patchData.Username
	}
	if patchData.ExpiryDate != 0 {
		update["expiry_date"] = patchData.ExpiryDate
	}
	if patchData.Password != "" {
		update["password"] = patchData.Password
	}
	if patchData.Outputs != nil {
		update["outputs"] = patchData.Outputs
	}

	// Perform the update on the MongoDB collection
	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error updating user",
		})
	}

	msg := "the user " + username + " has been updated successfully"

	return c.Status(200).JSON(fiber.Map{
		"status":  true,
		"message": msg,
	})
}

func helloWorld(c *fiber.Ctx) error {
	return c.JSON("Hello, World!")
}

func main() {
	// Connect to MongoDB
	initMongoDB()

	// Create a new Fiber app
	app := fiber.New()

	// Define routes and corresponding handlers
	app.Get("/", helloWorld)
	app.Post("/user", addUser)
	app.Get("/user/:username", getUserByUsername)
	app.Patch("/user/:username", updateUserByUsername)

	// Start the server on port 8080
	err := app.Listen(":8080")
	if err != nil {
		panic(err)
	}
}
