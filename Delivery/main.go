package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/routers"
	"github.com/Amaankaa/Blog-Starter-Project/Infrastructure"
	"github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/Amaankaa/Blog-Starter-Project/Usecases"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
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

	// ✅ Initialize email verifier
	emailVerifier, err := infrastructure.NewMailboxLayerVerifier()
	if err != nil {
		log.Fatalf("Failed to initialize email verifier: %v", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(userCollection, passwordService, emailVerifier)

	// Initialize usecases
	userUsecase := usecases.NewUserUsecase(userRepo)

	// Initialize controllers
	controller := controllers.NewController(userUsecase)

	// Setup router
	r := routers.SetupRouter(controller)

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}