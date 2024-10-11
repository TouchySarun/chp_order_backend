package models

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"testing"
	"time"
)

func stringPtr(s string) *string {
	return &s
}

func TestFilterOrderStringMatch(t *testing.T) {
	orders := []models.Order{
		{
			Id:        stringPtr("TaEDKpictzAAieRp9BNT"),
			Branch:    "000-DC",
			Name:      "สิงห์ เลมอนโชดา 330 มล ยูชุ(สัม)",
			UtqName:   "หีบ 24",
			UtqQty:    24,
			Code:      "1991000001267",
			Sku:       "fCYmGkURVbvSBHQeQq4H",
			Ap:        "1991-บริษัท บุญรอดเทรดดิ้ง จำกัด",
			Qty:       2,
			LeftQty:   2,
			Cat:       "3020-น้ำหวานต่างๆ",
			Bnd:       "0565-สิงห์",
			CreBy:     "TOUCH",
			StartDate: time.Now(),
			Status:    "init",
		},
		{
			Id:        stringPtr("b2w8UXDDOtRhkqMD7o0E"),
			Branch:    "000-DC",
			Name:      "สิงห์ เลม่อนโซดา 330 มล บ๊วย",
			UtqName:   "หีบ 24",
			UtqQty:    24,
			Code:      "8850999021232",
			Sku:       "ddVkINHkzf29yqylAat4",
			Ap:        "1991-บริษัท บุญรอดเทรดดิ้ง จำกัด",
			Qty:       2,
			LeftQty:   2,
			Cat:       "3040-น้ำอัดลม",
			Bnd:       "0565-สิงห์",
			CreBy:     "TOUCH",
			StartDate: time.Now(),
			Status:    "init",
		},
		{
			Id:        stringPtr("wCDUEcCVUGMS3EgDeqXw"),
			Branch:    "000-DC",
			Name:      "ซัมยัง แทงเกิ้ล บูลโกกิ 110 ก ซอสบาร์บีคิว+ครีมอัลเฟรโด้",
			UtqName:   "หีบ.32",
			UtqQty:    32,
			Code:      "10192699000076",
			Sku:       "EzoFVZG6Ndvnm2igBm5R",
			Ap:        "1217-บริษัท ซีโน-แปซิฟิค เทรดดิ้ง (ไทยแลนด์) จำกัด",
			Qty:       2,
			LeftQty:   2,
			Cat:       "1070-บะหมี่กึ่งสำเร็จรูป โจ๊กและซุป",
			Bnd:       "6743-ซัมยัง",
			CreBy:     "TOUCH",
			StartDate: time.Now(),
			Status:    "init",
		},
	}

	// Define the test cases
	tests := []struct {
		name   string
		mode   string
		match  string
		expect int // expected number of filtered orders
	}{
		{
			name:   "Test Matching Name Field with one keyword",
			mode:   "Name",
			match:  "สิง",
			expect: 2,
		},
		{
			name:   "Test Matching Name Field with multiple keywords",
			mode:   "Name",
			match:  "สิงห์ บ๊วย",
			expect: 1,
		},
		{
			name:   "Test Matching Name Field with no match",
			mode:   "Name",
			match:  "ไม่มี",
			expect: 0,
		},
		{
			name:   "Test Matching AP Field",
			mode:   "Ap",
			match:  "1991",
			expect: 2,
		},
	}

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services.FilterOrderStringMatch(orders, tt.mode, tt.match)
			if len(result) != tt.expect {
				t.Errorf("expected %d orders, but got %d", tt.expect, len(result))
			}
		})
	}
}