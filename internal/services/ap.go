package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"

	"google.golang.org/api/iterator"
)

func GetApsByRack (ctx context.Context, rack string) (*[]models.Ap, error){
	var aps []models.Ap
	query := firestore.Client.Collection("aps").Where("rack", "==", rack)
	iter := query.Documents(ctx)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Failed, Get aps from firestore: %v",err)
			return nil, fmt.Errorf("failed, Get aps from firestore %v",err)
		}
		var ap models.Ap
		if err := doc.DataTo(&ap); err != nil {
			fmt.Printf("Failed, convert apData to ap: %v", err)
			return nil, fmt.Errorf("failed, convert apData to ap %v", err)
		}
		aps = append(aps, ap)
	}



	return &aps,nil 
}