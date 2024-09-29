package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"log"

	"google.golang.org/api/iterator"
)

const usersCollection = "users"

func GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	doc, err := firestore.Client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
			return nil, err
	}
	return doc.Data(), nil
}

func GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var users []models.User
	query := firestore.Client.Collection(usersCollection).Where("username","==", username)
	iter := query.Documents(ctx)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed, Get user from firestore: %v",err)
			return nil, err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Fatalf("Failed, convert useData to user: %v", err)
			return nil, err
		}
		user.Id = &doc.Ref.ID
		users = append(users, user)
	}
	if len(users) == 0 {
		log.Fatalf("User not found: %v", username)
		return nil, fmt.Errorf("user not found: %v", username)
	}
	return &users[0], nil
}