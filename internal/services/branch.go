package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"context"
	"fmt"

	defFirestore "cloud.google.com/go/firestore"
)

func GetBranches(ctx context.Context) (*[]string, error) {
	var branchesMap map[string]string
	var branches []string
	doc ,err := firestore.Client.Collection("system").Doc("branch").Get(ctx)
	if err != nil {
		fmt.Printf("Failed, Getting branches: %v", err)
		return nil, err
	}
	if err := doc.DataTo(&branchesMap); err != nil {
		fmt.Printf("Failed, convert branchesData to branches: %v", err)
		return nil, err
	}

	for _, br  := range branchesMap {
		branches = append(branches, br)
	}
	return &branches, nil
}

func CreateBranch(ctx context.Context, code string, name string) error {
	docRef := firestore.Client.Collection("system").Doc("branch")

	newBranch := []defFirestore.Update{
		{
			Path: code,
			Value: name,
		},
	}
	_, err := docRef.Update(ctx, newBranch)
	return err
}