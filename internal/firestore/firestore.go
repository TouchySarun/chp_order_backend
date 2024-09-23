package firestore

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var Client *firestore.Client

// Initialize Firestore client
func InitFirestore() {
	ctx := context.Background()

	dir, _ := os.Getwd()
	saPath := filepath.Join(dir, "firebase-adminsdk.json")

	sa := option.WithCredentialsFile(saPath)

	var err error
	Client, err = firestore.NewClient(ctx, "test-5f1af", sa)
	if err != nil {
			log.Fatalf("Failed to create Firestore client: %v", err)
	}
}