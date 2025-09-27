package health

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client

func init() {
	// Try to load .env file, but don't fail if it doesn't exist
	// Environment variables can also be set directly
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors
		// This allows the application to work with environment variables set directly
	}
}

// func ConnectDB() error {
// 	uri := os.Getenv("MONGODB_URI")
// 	if uri == "" {
// 		return fmt.Errorf("MONGODB_URI environment variable not set")
// 	}

// 	clientOptions := options.Client().ApplyURI(uri)
// 	client, err := mongo.Connect(context.TODO(), clientOptions)
// 	if err != nil {
// 		return fmt.Errorf("unable to connect to MongoDB: %v", err)
// 	}

// 	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
// 		return fmt.Errorf("unable to ping MongoDB: %v", err)
// 	}

// 	Client = client
// 	return nil
// }

// ConnectDBWithTimeout creates a MongoDB connection with a timeout context using environment variables
func ConnectDBWithTimeout(timeout time.Duration) (*mongo.Client, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGODB_URI environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	
	// Add additional TLS configuration for Atlas
	clientOptions.SetTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS12,
	})
	
	clientOptions.SetServerSelectionTimeout(30 * time.Second)
	clientOptions.SetConnectTimeout(30 * time.Second)
	clientOptions.SetSocketTimeout(30 * time.Second)
	
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to MongoDB: %v", err)
	}

	// Test the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("unable to ping MongoDB: %v", err)
	}

	return client, nil
}

// GetCollection returns a MongoDB collection using environment variables for database and collection names
func GetCollection(client *mongo.Client) *mongo.Collection {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "iot" // default database name
	}
	
	collName := os.Getenv("COLL_NAME")
	if collName == "" {
		collName = "readings" // default collection name
	}
	
	return client.Database(dbName).Collection(collName)
}

