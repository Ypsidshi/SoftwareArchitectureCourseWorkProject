package repository

import "testing"

func TestIsExactInvoiceAmount(t *testing.T) {
	tests := []struct {
		name          string
		invoiceAmount float64
		paymentAmount float64
		want          bool
	}{
		{name: "exact match", invoiceAmount: 100, paymentAmount: 100, want: true},
		{name: "partial payment", invoiceAmount: 100, paymentAmount: 50, want: false},
		{name: "overpayment", invoiceAmount: 100, paymentAmount: 120, want: false},
		{name: "zero payment", invoiceAmount: 100, paymentAmount: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExactInvoiceAmount(tt.invoiceAmount, tt.paymentAmount); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
