package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	defFirestore "cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
)
const ordersCollection = "orders"


func EditAndCreateOrderHistory(ctx context.Context, id string, updatedFields map[string]interface{}, oh models.OrderHistory) error {
	// Start Firestore transaction to ensure consistency
	return firestore.Client.RunTransaction(ctx, func(ctx context.Context, tx *defFirestore.Transaction) error {
		// Edit order within the transaction
		if err := EditOrder(ctx, id, updatedFields); err != nil {
			return err
		}
		// Create order history within the transaction
		if err := CreateOrderHistory(ctx, id, oh); err != nil {
			return err
		}
		return nil
	})
}

func CreateOrderAndOrderHistory(ctx context.Context, o models.Order, oh models.OrderHistory) error {
	return firestore.Client.RunTransaction(ctx, func(ctx context.Context, tx *defFirestore.Transaction) error {
		// Edit order within the transaction
		id, err := CreateOrder(ctx, o)
		if err != nil {
			return err
		}
		// Create order history within the transaction
		if err := CreateOrderHistory(ctx, *id, oh); err != nil {
			return err
		}
		return nil
	})
}

func GetOrder(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	doc ,err := firestore.Client.Collection(ordersCollection).Doc(id).Get(ctx)
	if err != nil {
		fmt.Printf("Failed, Getting order: %v", err)
		return nil, err
	}
	if err := doc.DataTo(&order); err != nil {
		fmt.Printf("Failed, convert orderData to order: %v", err)
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
			fmt.Printf("Failed, Get sku from firestore: %v",err)
			return nil, err
		}
		var sku models.Sku
		if err := doc.DataTo(&sku); err != nil {
			fmt.Printf("Failed, convert skuData to sku: %v", err)
			return nil, err
		}
		sku.Id = &doc.Ref.ID
		skus = append(skus, sku)
	}
	if len(skus) == 0 {
		fmt.Printf("Barcode not found: %v", barcode)
		return nil, fmt.Errorf("barcode not found: %v", barcode)
	}
	return &skus[0], nil
}

func GetLatestOrder(ctx context.Context, skuId string, branch string) (*models.Order, error) {
	var orders []models.Order
	query := firestore.Client.Collection(ordersCollection).Where("leftQty", ">", 0).Where("sku", "==", skuId).Where("branch", "==", branch)
	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed getting data from iter \n%v", err)
		}
		var order models.Order
		if err := doc.DataTo(&order); err != nil {
			return nil, fmt.Errorf("failed convert data to order : %v \n%v", order, err)
		}
		order.Id = &doc.Ref.ID
		orders = append(orders, order)
	}
	if len(orders) > 0 {
		return &orders[0], nil
	} else {
		return nil, nil
	}
}	

func GetLatestSuccessOrderDate(ctx context.Context, skuId string, branch string) (*time.Time, error) {
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
	var defDate time.Time
	// have value
	if maxDate.After(defDate){
		return &maxDate, nil
	}else {
		return nil, nil
	}

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
func MakeBaseQueryForOrderPicking(status []string) defFirestore.Query {
	fmt.Printf("create base query with status: %v\n", status)
	today := time.Now().Truncate(24 * time.Hour)
	query := firestore.Client.Collection(ordersCollection).Where("leftQty", ">", 0).Where("status","in", status).Where("lstUpd", "<", today)
	return query
}
func FilterOrderStringMatch(orders []models.Order, mode string, match string) []models.Order {
	matches := strings.Split(match, " ")
	if matches[0] == "" {
		return orders
	}
	resOrder := make([]models.Order, 0)
	for _, o := range orders {
		val := reflect.ValueOf(o).FieldByName(mode).String()
		isAllMatch := true
		// if any of m dosn't match return false
		for _, m := range matches {
			isAllMatch = strings.Contains(val, m)
		}
		if isAllMatch {
			resOrder = append(resOrder, o)
		}
	}
	return resOrder
}
func filterOrderStringEqual(orders []models.Order, mode string, equal string) []models.Order {
	resOrder := make([]models.Order, 0)
	for _, o := range orders {
		val := reflect.ValueOf(o).FieldByName(mode).String()
		if val == equal {
			resOrder = append(resOrder, o)
		}
	}
	return resOrder
}
func filterArrayContain(orders []models.Order, mode string, match []string) []models.Order {
	resOrder := make([]models.Order, 0)
	for _, o := range orders {
		val := reflect.ValueOf(o).FieldByName(mode).String()
		if slices.Contains(match, val) {
			resOrder = append(resOrder, o)
		}
	}
	return resOrder
}
/**
* Get Orders For Picking
* @param q
* @return 1. orders:*[]models.Order
* @return 2. total orders: int
* @return 3. total orders before pagination: int
*/
func GetOrdersForPicking(ctx context.Context, q models.OrderQuery) (*[]models.Order, int, int, error){
	var orders []models.Order
	var query = MakeBaseQueryForOrderPicking(q.Status)
	var totalOrders int
	var totalOrdersBeforePagination int
	// get all orders (no filter, limit, offset, orderBy)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		fmt.Printf("fail get orders %v\n",err)
		return nil,0 ,0 , fmt.Errorf("failed get orders from firestore, %v",err)
	}
	for _, doc := range docs {
		var order models.Order 
		if err := doc.DataTo(&order); err == nil {
			order.Id = &doc.Ref.ID
			order.History = &[]models.OrderHistory{}
			orders = append(orders, order)
		}else {
			return nil,0 ,0 , fmt.Errorf("failed convert orderData to order, %v", err)
		}
	}
	totalOrders = len(orders)
	fmt.Printf("success get orders: [%v], start filtering\n", len(orders))
	if q.Ap != "" {
		fmt.Printf("filter Ap: %v\n",q.Ap)
		orders = FilterOrderStringMatch(orders, "Ap", q.Ap)
		fmt.Printf("complete filter Ap: [%v]\n", len(orders))
	}
	if len(q.ApContain) > 0 {
		fmt.Printf("filter Ap Contain: %v\n", q.ApContain)
		orders = filterArrayContain(orders,"ApContain", q.ApContain)
		fmt.Printf("complete filter Ap Contain: [%v]\n", len(orders))
	}
	if q.Bnd != "" {
		fmt.Printf("filter Bnd: %v\n",q.Bnd)
		orders = FilterOrderStringMatch(orders, "Bnd", q.Bnd)
		fmt.Printf("complete filter Bnd: [%v]\n", len(orders))
	}
	if len(q.Branch) > 0 {
		fmt.Printf("filter Branch: %v\n",q.Branch)
		orders = filterArrayContain(orders,"Branch", q.Branch)
		fmt.Printf("complete filter Branch: [%v]\n", len(orders))
	}
	if len(q.Code) > 0 {
		fmt.Printf("filter Code: %v\n",q.Code)
		orders = filterArrayContain(orders,"Code", q.Code)
		fmt.Printf("complete filter Code: [%v]\n", len(orders))
	}
	if q.CreBy != "" {
		fmt.Printf("filter CreBy: %v\n",q.CreBy)
		orders = filterOrderStringEqual(orders, "CreBy", q.CreBy)
		fmt.Printf("complete filter CreBy: [%v]\n", len(orders))
	}
	// rack was remove
	if q.Search != "" {
		fmt.Printf("filter Search: %v\n", q.Search)
		matchName := FilterOrderStringMatch(orders, "Name", q.Search)
		matchCode := FilterOrderStringMatch(orders, "Code", q.Search)
		orders = Union(matchName, matchCode)
		fmt.Printf("complete filter Search: [%v]\n", len(orders))
	}
	fmt.Printf("complete filter: [%v]\n", len(orders))
	// status is already query
	totalOrdersBeforePagination = len(orders)
	parts := strings.Split(q.OrderBy, "_")
	sortBy, mode := parts[0], strings.ToLower(parts[1])
	SortOrders(orders, sortBy, mode)
	// orders = ApplyLimitAndOffset(orders, q.Limit, q.Offset)
	return &orders, totalOrders, totalOrdersBeforePagination, nil
}
func UpdateStatus(ctx context.Context, id string, status string, qty int, creBy string) (*string, error){
	var updatedFields = map[string]interface{}{"lstUpd":time.Now()}
	order, oerr := GetOrder(ctx, id)
	if oerr != nil {
		// WriteResponseErr(&w, fmt.Sprintf("can't find order. \n %v\n",oerr), http.StatusBadRequest)
		return nil, fmt.Errorf("can not find order: %v",oerr)
	}
	var oh = models.OrderHistory{
		Date:       time.Now(),
		CreBy:      creBy,
	}
	// create updateFields and orderHistory
	switch status{
	case "picking":{
		// pc click order => just update status, lstUpd ## except status == "shipping"
		if order.Status == "shipping" || order.Status == "picking" {
			// WriteResponseSuccess(&w, fmt.Sprintf("Must update status to picking (PC click order) but order status is %v, so I do nothing.", order.Status))
			res := fmt.Sprintf("Must update status to picking (PC click order) but order status is %v, so I do nothing.", order.Status)
			return &res , nil
		} else {
			updatedFields["status"]="picking"
			oh.Status = "picking"
			break
		}
	}
	case "shipping":{
		// pc click picking done => check if (qty == 0) just update status to left else update status to shipping left qty
		if order.Status != "shipping" && order.Status != "picking" {
			// WriteResponseSuccess(&w, fmt.Sprintf("Must update status to shipping (PC done picking, order status shuld be picking or shipping) but order status is %v.", order.Status))
			res := fmt.Sprintf("Must update status to shipping (PC done picking, order status shuld be picking or shipping) but order status is %v.", order.Status)
			return &res, nil
		}
		if qty == 0 {
			// WriteResponseSuccess(&w, "Must update status to shipping and qty is 0, so I do nothing.")
			res := "Must update status to shipping and qty is 0, so I do nothing."
			return &res, nil
		} else {
			updatedFields["status"] = "shipping"
			updatedFields["leftQty"] = -1 * qty

			newQty := order.LeftQty - qty
			oh.Status = "shipping"
			oh.OldQty = &order.LeftQty
			oh.NewQty = &newQty
			break
		}
	}
	case "done": {
		// br click shipping done => check if (leftQty == 0) 
		if order.LeftQty == 0 {
			updatedFields["status"] = "done"
			updatedFields["endDate"] = time.Now()
			oh.Status = "done"
			break
		} else {
			updatedFields["status"] = "left"
			oh.Status = "left"
			break
		}
	}
	}
	if err := EditAndCreateOrderHistory(ctx,id,updatedFields,oh); err != nil {
		// WriteResponseErr(&w, fmt.Sprintf("Transaction failed: %v", err), http.StatusInternalServerError)
		return nil, fmt.Errorf("transaction failed: %v", err)
	} else {
		// WriteResponseSuccess(&w, "Success")
		res := "Success"
		return &res, nil
	}

}

func SortOrders(orders []models.Order, sortBy string, mode string) {
	switch sortBy {
	case "name":
		if mode == "asc" {
			sort.Slice(orders, func(i, j int) bool {
				return strings.ToLower(orders[i].Name) < strings.ToLower(orders[j].Name)
			})
		} else if mode == "desc" {
			sort.Slice(orders, func(i, j int) bool {
				return strings.ToLower(orders[i].Name) > strings.ToLower(orders[j].Name)
			})
		}
	case "code":
		if mode == "asc" {
			sort.Slice(orders, func(i, j int) bool {
				return strings.ToLower(orders[i].Code) < strings.ToLower(orders[j].Code)
			})
		} else if mode == "desc" {
			sort.Slice(orders, func(i, j int) bool {
				return strings.ToLower(orders[i].Code) > strings.ToLower(orders[j].Code)
			})
		}
	default:
		fmt.Println("Invalid sortBy parameter. Please use 'name' or 'code'.")
	}
}
