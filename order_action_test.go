package main

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	"github.com/stretchr/testify/require"
)

func validOrderOptions() createOrderOptions {
	return createOrderOptions{
		Consumer:         "shipper@example.org",
		Payer:            "payer@example.org",
		Purchaser:        "fake-purchaser@example.org",
		Items:            []string{"DBS0235559"},
		Courier:          "FedEx",
		CourierAccount:   "FAKE-ACCT-0001",
		Payment:          "credit_card",
		Comments:         "automated CLI test order",
		PurchaseOrderNum: "PO-FAKE-0001",
		Status:           string(StatusInPreparation),
	}
}

func TestParseStatusFP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StatusEnum
		wantErr bool
	}{
		{
			name:  "IN_PREPARATION status",
			input: "IN_PREPARATION",
			want:  StatusInPreparation,
		},
		{
			name:  "GROWING status",
			input: "GROWING",
			want:  StatusGrowing,
		},
		{
			name:  "CANCELLED status",
			input: "CANCELLED",
			want:  StatusCancelled,
		},
		{
			name:  "SHIPPED status",
			input: "SHIPPED",
			want:  StatusShipped,
		},
		{
			name:    "empty status is invalid",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid status",
			input:   "LOST",
			wantErr: true,
		},
		{
			name:    "lowercase status is invalid",
			input:   "shipped",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStatus(tt.input)
			val, err := E.Unwrap(result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, val)
		})
	}
}

func TestBuildCreateOrderInput(t *testing.T) {
	opts := validOrderOptions()
	result := buildCreateOrderInput(opts)
	val, err := E.Unwrap(result)
	require.NoError(t, err)
	require.Equal(t, opts.Consumer, val.Consumer)
	require.Equal(t, opts.Payer, val.Payer)
	require.Equal(t, opts.Purchaser, val.Purchaser)
	require.Equal(t, opts.Items, val.Items)
	require.Equal(t, opts.Courier, val.Courier)
	require.Equal(t, opts.CourierAccount, val.CourierAccount)
	require.Equal(t, opts.Payment, val.Payment)
	require.Equal(t, opts.Comments, val.Comments)
	require.Equal(t, opts.PurchaseOrderNum, val.PurchaseOrderNum)
	require.Equal(t, StatusInPreparation, val.Status)
	require.NotNil(t, val.ConsumerInfo, "should attach fake consumer info")
	require.NotNil(t, val.PayerInfo, "should attach fake payer info")
}

func TestBuildCreateOrderInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*createOrderOptions)
		wantErr string
	}{
		{
			name:    "missing consumer email",
			mutate:  func(o *createOrderOptions) { o.Consumer = "not-an-email" },
			wantErr: "consumer",
		},
		{
			name:    "missing payer email",
			mutate:  func(o *createOrderOptions) { o.Payer = "" },
			wantErr: "payer",
		},
		{
			name:    "missing purchaser email",
			mutate:  func(o *createOrderOptions) { o.Purchaser = "buyer" },
			wantErr: "purchaser",
		},
		{
			name:    "no stock items",
			mutate:  func(o *createOrderOptions) { o.Items = nil },
			wantErr: "item",
		},
		{
			name:    "invalid status",
			mutate:  func(o *createOrderOptions) { o.Status = "LOST" },
			wantErr: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validOrderOptions()
			tt.mutate(&opts)
			result := buildCreateOrderInput(opts)
			_, err := E.Unwrap(result)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
