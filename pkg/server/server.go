package server

import (
	"test_piece/pkg/controllers"
	"test_piece/pkg/db"

	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", controllers.Home)
	app.Post("/user", controllers.AddUser)
	app.Get("/users", controllers.GetUsers)
	app.Get("/user/:username", controllers.GetUserByUsername)
	app.Patch("/user/:username", controllers.UpdateUserByUsername)
}

func Run() {
	db.InitMongoDB()

	app := fiber.New()

	setupRoutes(app)

	err := app.Listen(":8080")
	if err != nil {
		panic(err)
	}
}
