package models

import (
	"encoding/json"
	"testing"
)

func TestCurrencyRate_JSONSerialization(t *testing.T) {
	// Test that CurrencyRate serializes to JSON correctly
	rate := CurrencyRate{
		From: "USD",
		To:   "EUR",
		Rate: 0.85,
		Date: "2024-01-15",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(rate)
	if err != nil {
		t.Fatalf("Failed to marshal CurrencyRate to JSON: %v", err)
	}

	// Verify JSON structure
	expected := `{"from":"USD","to":"EUR","rate":0.85,"date":"2024-01-15"}`
	if string(jsonData) != expected {
		t.Errorf("JSON serialization mismatch.\nExpected: %s\nActual: %s", expected, string(jsonData))
	}

	// Test unmarshaling back
	var unmarshaled CurrencyRate
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to CurrencyRate: %v", err)
	}

	// Verify all fields match
	if unmarshaled.From != rate.From {
		t.Errorf("From field mismatch: expected %s, got %s", rate.From, unmarshaled.From)
	}
	if unmarshaled.To != rate.To {
		t.Errorf("To field mismatch: expected %s, got %s", rate.To, unmarshaled.To)
	}
	if unmarshaled.Rate != rate.Rate {
		t.Errorf("Rate field mismatch: expected %f, got %f", rate.Rate, unmarshaled.Rate)
	}
	if unmarshaled.Date != rate.Date {
		t.Errorf("Date field mismatch: expected %s, got %s", rate.Date, unmarshaled.Date)
	}
}

func TestConvertResponse_JSONSerialization(t *testing.T) {
	// Test that ConvertResponse serializes to JSON correctly
	response := ConvertResponse{
		Amount: 123.45,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ConvertResponse to JSON: %v", err)
	}

	// Verify JSON structure
	expected := `{"amount":123.45}`
	if string(jsonData) != expected {
		t.Errorf("JSON serialization mismatch.\nExpected: %s\nActual: %s", expected, string(jsonData))
	}

	// Test unmarshaling back
	var unmarshaled ConvertResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to ConvertResponse: %v", err)
	}

	// Verify amount matches
	if unmarshaled.Amount != response.Amount {
		t.Errorf("Amount field mismatch: expected %f, got %f", response.Amount, unmarshaled.Amount)
	}
}

func TestCurrencyRate_ZeroValues(t *testing.T) {
	// Test behavior with zero/empty values
	rate := CurrencyRate{}

	jsonData, err := json.Marshal(rate)
	if err != nil {
		t.Fatalf("Failed to marshal empty CurrencyRate: %v", err)
	}

	expected := `{"from":"","to":"","rate":0,"date":""}`
	if string(jsonData) != expected {
		t.Errorf("Empty CurrencyRate JSON mismatch.\nExpected: %s\nActual: %s", expected, string(jsonData))
	}
}
