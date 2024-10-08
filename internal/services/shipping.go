package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"sort"
	"time"
)
const shippingCollection = "shipping"

func CreateShipping (ctx context.Context, shipping models.Shipping) error {
	_, _, err := firestore.Client.Collection(shippingCollection).Add(ctx, shipping)
	return err 
}

func GetShipping(ctx context.Context, branch string, limit int, page int) (*[]models.Shipping, error) {
	var shippings []models.Shipping
	var shippingsMap = make(map[string]models.Shipping)
	offset := (page - 1) * limit

	// Query shipping records by branch
	query := firestore.Client.Collection(shippingCollection).Where("branch", "==", branch)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	// Loop through documents and aggregate shipping quantities by OrderId
	for _, doc := range docs {
		var shipping models.Shipping
		if err := doc.DataTo(&shipping); err == nil {
			// Handle nil case for shippingsMap[shipping.OrderId]
			existingShipping, found := shippingsMap[shipping.OrderId]
			if !found {
				existingShipping.Qty = 0 // Treat as 0 if not found
			}
			
			// Aggregate quantity
			newQty := shipping.Qty + existingShipping.Qty
			shippingsMap[shipping.OrderId] = models.Shipping{
				Id:      doc.Ref.ID,
				OrderId: shipping.OrderId,
				Qty:     newQty,
				Branch:  shipping.Branch,
			}
		}
	}

	// Collect the aggregated shippings into a slice
	for _, shipping := range shippingsMap {
		shippings = append(shippings, shipping)
	}

	// Sort the slice by OrderId in ascending order
	sort.Slice(shippings, func(i, j int) bool {
		return shippings[i].OrderId < shippings[j].OrderId
	})

	// Apply limit and offset for pagination
	start := offset
	end := offset + limit
	if start > len(shippings) {
		start = len(shippings) // Prevent index out of range
	}
	if end > len(shippings) {
		end = len(shippings)
	}
	paginatedShippings := shippings[start:end]

	return &paginatedShippings, nil
}

func GetDefectShipping(ctx context.Context, branch string, limit int, page int) {
	// TODO:
}

func ConfirmShipping(ctx context.Context, branch string, creBy string) error {
	// Initialize the map for aggregating shipping quantities by OrderId
	shippingsMap := make(map[string]models.Shipping)

	// Query shipping records by branch
	query := firestore.Client.Collection(shippingCollection).Where("branch", "==", branch)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	// Loop through documents and aggregate shipping quantities by OrderId
	for _, doc := range docs {
		var shipping models.Shipping
		if err := doc.DataTo(&shipping); err == nil {
			// Aggregate quantity, handle nil case for shippingsMap[shipping.OrderId]
			existingShipping := shippingsMap[shipping.OrderId]
			newQty := shipping.Qty + existingShipping.Qty // Treat as 0 if not found
			shippingsMap[shipping.OrderId] = models.Shipping{
				Id:      doc.Ref.ID,
				OrderId: shipping.OrderId,
				Qty:     newQty,
				Branch:  shipping.Branch,
			}
		}
	}

	// Process each aggregated shipping by OrderId
	for id, shipping := range shippingsMap {
		order, err := GetOrder(ctx, id)
		if err != nil {
			return err
		}

		// Handle case where order.History is nil
		if order.History == nil {
			return fmt.Errorf("order history is missing for OrderId: %s", id)
		}

		// Get the latest shipping quantity from the order history
		dcQty := getLatestShippingQty(*order.History)

		// If the aggregated shipping quantity matches the historical quantity, update the status
		if dcQty == shipping.Qty {
			// Set status to "done"
			if _, err := UpdateStatus(ctx, id, "done", 1, creBy); err != nil {
				return err
			}
		} else {
			// Adjust the current leftQty and update the order
			// qty:100, leftQty:90 -> shipping(50) -> confirm(25) -> leftQty = now 40 => 65 , change +25(shipping-confirm)
			// qty:100, leftQty:90 -> shipping(50) -> confirm(75) -> leftQty = now 40 => 15 , change -25(shipping-confirm)
			dif := dcQty - shipping.Qty
			updatedFields := map[string]interface{}{
				"leftQty": dif,
			}

			// Update the order with the adjusted leftQty
			if err := EditOrder(ctx, id, updatedFields); err != nil {
				return err
			}

			// Set status to "done"
			if _, err := UpdateStatus(ctx, id, "done", 1, creBy); err != nil {
				return err
			}
		}
	}

	return nil
}

func getLatestShippingQty (oh []models.OrderHistory) int {
	var flag = false
	var maxDate time.Time
	var qty int

	for _, h := range oh {

		if !flag || (h.Date.Before(maxDate) && h.Status=="shipping") {
			maxDate = h.Date
			qty = *h.OldQty - *h.NewQty
		}
	}
	
	return qty
}