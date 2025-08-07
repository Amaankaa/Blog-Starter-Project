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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

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

	db := client.Database("blog_db")
	userCollection := db.Collection("users")
	tokenCollection := db.Collection("tokens")
	blogCollection := db.Collection("blogs")
	passwordResetCollection := db.Collection("password_resets")

	// Initialize infrastructure services
	passwordService := infrastructure.NewPasswordService()
	jwtService := infrastructure.NewJWTService()

	emailVerifier, err := infrastructure.NewMailboxLayerVerifier()
	if err != nil {
		log.Fatalf("Failed to initialize email verifier: %v", err)
	}
	emailSender := infrastructure.NewBrevoEmailSender()
	
	//Repositories: only take collection (not services)
	userRepo := repositories.NewUserRepository(userCollection)
	tokenRepo := repositories.NewTokenRepository(tokenCollection)
	blogRepo := repositories.NewBlogRepository(blogCollection)
	passwordResetRepo := repositories.NewPasswordResetRepo(passwordResetCollection, userCollection)
	//AI configuration
	aiAPIKey := os.Getenv("GEMINI_API_KEY")
	if aiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY not set in environment")
	}
	aiAPIURL := os.Getenv("GEMINI_API_URL")
	if aiAPIURL == "" {
		log.Fatal("GEMINI_API_URL not set in environment")
	}		

	//Usecase: handles business logic, gets all dependencies
	userUsecase := usecases.NewUserUsecase(
	userRepo,
	passwordService,
	tokenRepo,
	jwtService,
	emailVerifier,
	emailSender,
	passwordResetRepo,
)
	blogUsecase := usecases.NewBlogUsecase(blogRepo)
	aiUseCase := usecases.NewAIUseCase(aiAPIKey, aiAPIURL)
	//Controller
	controller := controllers.NewController(userUsecase)
	blogController := controllers.NewBlogController(blogUsecase)
	aiController := controllers.NewAIController(aiUseCase)
	// Initialize AuthMiddleware
	authMiddleware := infrastructure.NewAuthMiddleware(jwtService)
	aiRateLimiter := infrastructure.NewRateLimiter(infrastructure.RateLimit, infrastructure.BurstLimit)
	//Router
	r := routers.SetupRouter(controller, blogController, authMiddleware, aiController,aiRateLimiter)

	//Start Server
	log.Println("Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
