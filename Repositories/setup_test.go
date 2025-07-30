// test_setup_test.go
package repositories

import (
    "context"
    "log"
    "os"
    "testing"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var testMongoClient *mongo.Client

func TestMain(m *testing.M) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    testMongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatalf("Could not connect to MongoDB for tests: %v", err)
    }

    code := m.Run()

    _ = testMongoClient.Disconnect(ctx)
    os.Exit(code)
}
