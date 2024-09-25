package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"log"
	"time"

	defFirestore "cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
)

func GetSkuByBarcode(ctx context.Context, barcode string) (*models.Sku, error) {
	var skus []models.Sku
	query := firestore.Client.Collection("skus").Where("barcodes","array-contains", barcode)
	iter := query.Documents(ctx)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed, Get sku from firestore: %v",err)
			return nil, err
		}
		var sku models.Sku
		if err := doc.DataTo(&sku); err != nil {
			log.Fatalf("Failed, convert skuData to sku: %v", err)
			return nil, err
		}
		sku.Id = &doc.Ref.ID
		skus = append(skus, sku)
	}
	if len(skus) == 0 {
		log.Fatalf("Barcode not found: %v", barcode)
		return nil, fmt.Errorf("barcode not found: %v", barcode)
	}
	return &skus[0], nil
}

func GetLatestOrder(ctx context.Context, skuId string, branch string) (*models.Order, error) {
	var orders []models.Order
	query := firestore.Client.Collection("orders").Where("leftQty", ">", 0).Where("sku", "==", skuId)
	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if len(orders) > 0 {
		return &orders[0], nil
	} else {
		return nil, nil
	}
}	

func GetLatestSuccessOrderDate(ctx context.Context, skuId string, branch string) (*string, error) {
	var maxDate time.Time
	var found bool

	query := firestore.Client.Collection("orders").Where("leftQty", "==", 0).Where("sku", "==", skuId)
	iter := query.Documents(ctx)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			return nil, err
		}
		if order.EndDate == nil {
			return nil, fmt.Errorf("some thing wrong in db leftqty ==0 but not have enddate, orderId: %v", order.Id)
		}
		
		if !found || order.EndDate.Before(maxDate) {
			maxDate = *order.EndDate
			found = true
		}
	}
	lstDateStr := maxDate.String()

	return &lstDateStr, nil
}	

func CreateOrder (ctx context.Context, order models.Order) (*string, error) {
	docRef, _, err := firestore.Client.Collection("orders").Add(ctx, order)
	
	if err != nil {
		return nil, err
	}
	return &docRef.ID, nil
}

func CreateOrderHistory (ctx context.Context, orderId string, orderHistory models.OrderHistory) (*string, error) {
	docRef := firestore.Client.Collection("orders").Doc(orderId)
	
	_, err := docRef.Update(ctx, []defFirestore.Update{
		{
			Path: "history",
			Value: defFirestore.ArrayUnion(orderHistory),
		},
	})
	if err != nil {
		return nil, err
	}
	return &docRef.ID, nil
} 