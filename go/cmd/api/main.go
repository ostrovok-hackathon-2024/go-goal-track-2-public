package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/go-goal/tagger/internal/api"
)

func main() {
	// Create a new Fiber app
	app := fiber.New()

	// Setup routes
	api.SetupRoutes(app)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting server on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
