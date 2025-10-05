package models

import (
	"strings"
	"testing"
)

func TestProduct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		product Product
		wantErr bool
	}{
		{
			name: "valid product",
			product: Product{
				ProductID:    1,
				SKU:          "ABC123",
				Manufacturer: "Test Manufacturer",
				CategoryID:   1,
				Weight:       100,
				SomeOtherID:  1,
			},
			wantErr: false,
		},
		{
			name: "invalid product_id (zero)",
			product: Product{
				ProductID:    0,
				SKU:          "ABC123",
				Manufacturer: "Test Manufacturer",
				CategoryID:   1,
				Weight:       100,
				SomeOtherID:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid SKU (empty)",
			product: Product{
				ProductID:    1,
				SKU:          "",
				Manufacturer: "Test Manufacturer",
				CategoryID:   1,
				Weight:       100,
				SomeOtherID:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid SKU (too long)",
			product: Product{
				ProductID:    1,
				SKU:          strings.Repeat("a", 101), // 101 characters
				Manufacturer: "Test Manufacturer",
				CategoryID:   1,
				Weight:       100,
				SomeOtherID:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid weight (negative)",
			product: Product{
				ProductID:    1,
				SKU:          "ABC123",
				Manufacturer: "Test Manufacturer",
				CategoryID:   1,
				Weight:       -1,
				SomeOtherID:  1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Product.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProductID(t *testing.T) {
	tests := []struct {
		name    string
		id      int32
		wantErr bool
	}{
		{"valid ID", 1, false},
		{"valid ID large", 999999, false},
		{"invalid ID zero", 0, true},
		{"invalid ID negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProductID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProductID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
