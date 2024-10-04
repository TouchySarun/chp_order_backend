package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"fmt"
	"reflect"

	defFirestore "cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
)

const usersCollection = "users"

func GetUser(ctx context.Context, userID string) (*models.User, error) {
	doc, err := firestore.Client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		fmt.Printf("Failed, get data from firestore: %v", err)
		return nil, err
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		fmt.Printf("Failed, convert useData to user: %v", err)
		return nil, err
	}
	user.Id = userID
	return &user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var users []models.User
	query := firestore.Client.Collection(usersCollection).Where("username","==", username)
	iter := query.Documents(ctx)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			fmt.Printf("Failed, Get user from firestore: %v",err)
			return nil, err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			fmt.Printf("Failed, convert useData to user: %v", err)
			return nil, err
		}
		user.Id = doc.Ref.ID
		if user.Ap == nil {
			user.Ap = []string{}
		}
		// fmt.Printf("user %v\n",user)
		users = append(users, user)
	}
	if len(users) == 0 {
		fmt.Printf("User not found: %v", username)
		return nil, fmt.Errorf("user not found: %v", username)
	}
	return &users[0], nil
}

func CreateUser(ctx context.Context, user models.User) error {
	_, _, err := firestore.Client.Collection("users").Add(ctx, user)
	return err 
}

func EditUser(ctx context.Context, id string, user models.User) error {
	userRef := firestore.Client.Collection("users").Doc(id)
	updateFileds, err := MakeUserUpdateField(user)
	if err != nil {
		return err
	}
	_, err = userRef.Update(ctx, *updateFileds)
	return err
}

func MakeUserUpdateField (req models.User) (*[]defFirestore.Update , error) {
	reqValue := reflect.ValueOf(req) // get array of value from req
	reqType := reflect.TypeOf(req) // get array of type from req
	var updates []defFirestore.Update

	if reqValue.NumField() == 0 {
		return nil, fmt.Errorf("no input fields")
	}

	for i:=0; i< reqValue.NumField(); i++ {
		v := reqValue.Field(i) // get req[i].value
		t	:= reqType.Field(i) // get req[i].type
		n := t.Tag.Get("json") // get type name
		var newUpdate defFirestore.Update

		// if have value add to update fields
		switch v.Kind() {
			case reflect.String: // Handle strings (value type)
				if v.String() != "" { // Check if string is not empty
					newUpdate = defFirestore.Update{
						Path: n,
						Value: v.String(),
					}
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Handle integer types
				if v.Int() != 0 { // Check if integer is non-zero
					newUpdate = defFirestore.Update{
						Path: n,
						Value: v.Int(),
					}
				}
			case reflect.Slice: // Handle slices
				if v.Len() > 0 { // Only add non-empty slices
					sliceValues := make([]interface{}, v.Len())
					for j := 0; j < v.Len(); j++ {
						sliceValues[j] = v.Index(j).Interface() // Extract slice elements
					}
					newUpdate = defFirestore.Update{
						Path: n,
						Value: sliceValues,
					}
				}
		}
		if newUpdate.Path != "" {
			updates = append(updates, newUpdate)
		}
	}

	return &updates, nil
}