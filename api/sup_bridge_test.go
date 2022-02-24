package api

import (
	"testing"
)

func TestFetchETHPrice(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"fetch", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchETHPrice()
			if err != nil {
				t.Error(err)
			}
			t.Log(got.String())
		})
	}
}

func TestFetchBNBPrice(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"fetch", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchBNBPrice()
			if err != nil {
				t.Error(err)
			}
			t.Log(got.String())
		})
	}
}
