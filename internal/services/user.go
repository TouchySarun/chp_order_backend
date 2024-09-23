package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"context"
)

func GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	doc, err := firestore.Client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
			return nil, err
	}
	return doc.Data(), nil
}