package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	defFirestore "cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
)
const ordersCollection = "orders"
func GetOrder(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	doc ,err := firestore.Client.Collection(ordersCollection).Doc(id).Get(ctx)
	if err != nil {
		log.Fatalf("Failed, Getting order: %v", err)
		return nil, err
	}
	if err := doc.DataTo(&order); err != nil {
		log.Fatalf("Failed, convert orderData to order: %v", err)
		return nil, err
	}
	order.Id = &doc.Ref.ID
	return &order, nil
	
}

func EditOrder(ctx context.Context, id string, updatedFields map[string]interface{}) error {

	// Reference the specific order document
	orderRef := firestore.Client.Collection(ordersCollection).Doc(id)

	// Update the document with the fields provided in updatedFields
	_, err := orderRef.Update(ctx, updatedFieldsToFirestoreUpdates(updatedFields))
	if err != nil {
		return fmt.Errorf("failed to update order: %v", err)
	}

	return nil
}
func updatedFieldsToFirestoreUpdates(updatedFields map[string]interface{}) []defFirestore.Update {
	var updates []defFirestore.Update
	for field, value := range updatedFields {
		newUpdate := defFirestore.Update{
			Path: field,
			Value: value,
		}
		if field == "qty" || field == "leftQty" {
			newUpdate.Value = defFirestore.Increment(value)
		}
		updates = append(updates, newUpdate)
	}
	return updates
}
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
	query := firestore.Client.Collection(ordersCollection).Where("leftQty", ">", 0).Where("sku", "==", skuId)
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

	query := firestore.Client.Collection(ordersCollection).Where("leftQty", "==", 0).Where("sku", "==", skuId)
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
	docRef, _, err := firestore.Client.Collection(ordersCollection).Add(ctx, order)
	
	if err != nil {
		return nil, err
	}
	return &docRef.ID, nil
}

func CreateOrderHistory (ctx context.Context, orderId string, orderHistory models.OrderHistory) error {
	docRef := firestore.Client.Collection(ordersCollection).Doc(orderId)
	
	_, err := docRef.Update(ctx, []defFirestore.Update{
		{
			Path: "history",
			Value: defFirestore.ArrayUnion(orderHistory),
		},
	})
	if err != nil {
		return err
	}
	return nil
} 


func MakeOrderHistoryUpdateField (req models.OrderEditRequest, order models.Order) (models.OrderHistory){
	var newQty = order.Qty + req.Qty
	var newUtqName = order.UtqName
	if req.UtqName != "" {
		newUtqName = req.UtqName
	}
	orderHistory := models.OrderHistory{
		Status:     "edit",
		Date:       time.Now(),
		CreBy:      req.CreBy,
		OldQty:     &order.Qty,
		OldUtqName: &order.UtqName,
		NewQty:     &newQty,
		NewUtqName: &newUtqName,
	}
	return orderHistory
}

func MakeOrderUpdateField (req models.OrderEditRequest) (*map[string]interface{}, error) {
	updatedFields := make(map[string]interface{})
	reqValue := reflect.ValueOf(req) // get array of value from req
	reqType := reflect.TypeOf(req) // get array of type from req

	if reqValue.NumField() == 0 {
		return nil, fmt.Errorf("no input fields")
	}

	for i:=0; i< reqValue.NumField(); i++ {
		v := reqValue.Field(i) // get req[i].value
		t	:= reqType.Field(i) // get req[i].type
		n := t.Tag.Get("json") // get type name

		// if have value add to update fields
		switch v.Kind() {
			case reflect.String: // Handle strings (value type)
				if v.String() != "" { // Check if string is not empty
					updatedFields[n] = v.String()
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Handle integer types
				if v.Int() != 0 { // Check if integer is non-zero
					updatedFields[n] = v.Int()
				}
		}
	}

	if len(updatedFields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	if updatedFields["qty"] != nil {
		updatedFields["leftQty"] = req.Qty
	}
	updatedFields["lstUpd"] = time.Now()
	return &updatedFields, nil
}