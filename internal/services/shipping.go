package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"sort"
)
const shippingCollection = "shipping"

func CreateShipping (ctx context.Context, shipping models.Shipping) error {
	_, _, err := firestore.Client.Collection(shippingCollection).Add(ctx, shipping)
	
	if err != nil {
		return err
	}

	return nil
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