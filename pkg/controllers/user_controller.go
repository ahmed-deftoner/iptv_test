package controllers

import (
	"context"
	"math"
	"strconv"
	"strings"
	"test_piece/pkg/db"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRequest struct {
	Username   string   `bson:"username" json:"username"`
	ExpiryDate int64    `bson:"expiry_date" json:"expiry_date"`
	Outputs    []string `bson:"outputs" json:"outputs"`
	Password   string   `bson:"password" json:"password"`
}

func AddUser(c *fiber.Ctx) error {
	user := new(UserRequest)

	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request format",
		})
	}

	// Check if the expiry date is valid

	// expiryTime := time.Unix(user.Expiry, 0)
	// if expiryTime.Before(time.Now()) {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"message": "Expiry date should be in the future",
	// 	})
	// }

	if user.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "username required",
		})
	}
	if user.ExpiryDate == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "expiry date required",
		})
	}
	if len(user.Outputs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "outputs required",
		})
	}
	if user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "password required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingUser := db.User{}
	err := db.Collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  false,
			"message": "Username is not available",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error checking username availability",
		})
	}

	_, err = db.Collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error adding user",
		})
	}

	return c.Status(201).JSON(user)
}

func GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Username not provided",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"username": username}

	var user db.User
	err := db.Collection.FindOne(ctx, filter).Decode(&user)
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

func UpdateUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Username not provided",
		})
	}

	patchData := new(UserPatch)

	if err := c.BodyParser(patchData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request format",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"username": username}

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

	_, err := db.Collection.UpdateOne(ctx, filter, bson.M{"$set": update})
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

func GetUsers(c *fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "50"))
	if err != nil || limit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid 'limit' parameter",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid 'page' parameter",
		})
	}

	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	username := c.Query("username")
	if username != "" {
		filter["username"] = username
	}

	expiryStr := c.Query("expiry_date")
	if expiryStr != "" {
		expiry, err := strconv.ParseInt(expiryStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  false,
				"message": "Invalid 'expiry' parameter",
			})
		}
		filter["expiry_date"] = expiry
	}

	outputType := c.Query("outputs")
	if outputType != "" {
		filter["outputs"] = outputType
	}

	sortby := c.Query("sortBy")
	sortOrder := 1
	if order := c.Query("order"); strings.ToLower(order) == "desc" {
		sortOrder = -1
	}
	if strings.HasPrefix(sortby, "-") {
		sortOrder = -1
		sortby = strings.TrimPrefix(sortby, "-")
	}
	sortField := strings.ToLower(sortby)
	sortOptions := options.Find().SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	totalRecords, err := db.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting records",
		})
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	cursor, err := db.Collection.Find(ctx, filter, sortOptions.SetLimit(int64(limit)).SetSkip(int64(offset)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error retrieving users",
		})
	}
	defer cursor.Close(ctx)

	var users []db.User
	if err := cursor.All(ctx, &users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Error decoding users",
		})
	}

	response := fiber.Map{
		"records":       users,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return c.JSON(response)
}

func Home(c *fiber.Ctx) error {
	return c.JSON("Welcome")
}
