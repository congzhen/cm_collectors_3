package datatype

import (
	"encoding/json"
	"testing"
)

func TestCustomDateUnmarshalJSON(t *testing.T) {
	var withNull struct {
		LastScraperUpdateTime *CustomDate `json:"lastScraperUpdateTime"`
	}
	if err := json.Unmarshal([]byte(`{"lastScraperUpdateTime":null}`), &withNull); err != nil {
		t.Fatalf("unmarshal null CustomDate failed: %v", err)
	}
	if withNull.LastScraperUpdateTime != nil {
		t.Fatalf("expected nil CustomDate for null, got %#v", withNull.LastScraperUpdateTime)
	}

	var withDate struct {
		LastScraperUpdateTime *CustomDate `json:"lastScraperUpdateTime"`
	}
	if err := json.Unmarshal([]byte(`{"lastScraperUpdateTime":"2026-04-29"}`), &withDate); err != nil {
		t.Fatalf("unmarshal string CustomDate failed: %v", err)
	}
	if withDate.LastScraperUpdateTime == nil || withDate.LastScraperUpdateTime.IsZero() {
		t.Fatalf("expected non-zero CustomDate, got %#v", withDate.LastScraperUpdateTime)
	}
}

func TestCustomTimeUnmarshalJSON(t *testing.T) {
	var data struct {
		CreatedAt *CustomTime `json:"createdAt"`
	}
	if err := json.Unmarshal([]byte(`{"createdAt":"2026-04-29 08:20:07"}`), &data); err != nil {
		t.Fatalf("unmarshal string CustomTime failed: %v", err)
	}
	if data.CreatedAt == nil || data.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero CustomTime, got %#v", data.CreatedAt)
	}
}
