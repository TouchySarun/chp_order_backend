package services

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func CreateSku (ctx context.Context, sku models.Sku) (*string, error) {
	docRef, _, err := firestore.Client.Collection("skus").Add(ctx, sku)
	
	if err != nil {
		return nil, err
	}
	return &docRef.ID, nil
}



func ReadSkus(filename string) ([]models.Sku, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all rows from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	// Create a slice to store Sku data
	var skus []models.Sku
	var skusMap = make(map[string]models.Sku)

	// Loop through the CSV records, skipping the header (assuming first row is header)
	for i, record := range records {
		if i == 0 {
			// Skip header row
			continue
		}
		// Convert string to float64
		floatValue, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			return nil, fmt.Errorf("failed convert record[7] to int %v", record[7])
		}
		utqQty := int(floatValue)
		
		goods := models.Goods{
			Code: record[5],
			UtqName: record[6],
			UtqQty: utqQty,
			Price0: record[8],
			Prict8: record[9],
		}
		if sku, ok := skusMap[record[0]]; ok {
			sku.Goods = append(sku.Goods, goods)
			sku.Barcodes = append(sku.Barcodes, goods.Code)
			skusMap[record[0]] = sku
		} else {
			skusMap[record[0]] = models.Sku{
				Name:  record[0],
				Ap:  record[1],
				Img: record[2],
				Cat: record[3],
				Bnd: record[4],
				Barcodes: []string{goods.Code},
				Goods: []models.Goods{goods},
			}
		} 
	}
	for _, sku := range skusMap {
		skus = append(skus, sku)
	}

	return skus, nil
}