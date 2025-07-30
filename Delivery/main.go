package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/routers"
	infrastructure "github.com/Amaankaa/Blog-Starter-Project/Infrastructure"
	"github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/Amaankaa/Blog-Starter-Project/Usecases"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI from environment variable
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set in environment")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Select database and collections
	db := client.Database("blog_db")
	userCollection := db.Collection("users")

	// Initialize services
	passwordService := infrastructure.NewPasswordService()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(userCollection, passwordService)

	// Initialize usecases
	userUsecase := usecases.NewUserUsecase(userRepo)

	// Initialize controllers
	controller := controllers.NewController(userUsecase)

	// Initialize middleware

	// Setup router
	r := routers.SetupRouter(controller)

	// Start server
	log.Println("Server starting on :8080")
	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
