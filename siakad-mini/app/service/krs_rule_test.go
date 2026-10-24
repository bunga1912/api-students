package service

import "testing"

func TestMaxSKSByIPK(t *testing.T) {
	tests := []struct {
		ipk  float64
		want int
	}{
		{4.00, 24},
		{3.50, 24},
		{3.00, 24},
		{2.99, 21},
		{2.50, 21},
		{2.49, 18},
		{0, 18},
	}

	for _, tt := range tests {
		if got := MaxSKSByIPK(tt.ipk); got != tt.want {
			t.Errorf("IPK %.2f: dapat %d, mau %d", tt.ipk, got, tt.want)
		}
	}
}

func TestCanTakeSKS(t *testing.T) {
	tests := []struct {
		name       string
		ipk        float64
		current    int
		additional int
		want       bool
	}{
		{"pas di batas", 3.20, 21, 3, true},
		{"lewat 1 SKS", 3.20, 22, 3, false},
		{"IPK menengah, pas 21", 2.80, 18, 3, true},
		{"IPK menengah, lewat", 2.80, 19, 3, false},
		{"IPK rendah, pas 18", 2.10, 15, 3, true},
		{"IPK rendah, lewat", 2.10, 16, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanTakeSKS(tt.ipk, tt.current, tt.additional)
			if got != tt.want {
				t.Errorf("dapat %v, mau %v", got, tt.want)
			}
		})
	}
}