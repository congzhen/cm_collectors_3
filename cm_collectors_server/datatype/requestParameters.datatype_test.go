package datatype

import (
	"encoding/json"
	"testing"
)

func TestResourceRequestDistinguishesMissingAndEmptyDramaSeries(t *testing.T) {
	var missing ReqParam_Resource
	if err := json.Unmarshal([]byte(`{"resource":{}}`), &missing); err != nil {
		t.Fatalf("unmarshal missing dramaSeries: %v", err)
	}
	if missing.DramaSeries != nil {
		t.Fatalf("missing dramaSeries should remain nil: %#v", missing.DramaSeries)
	}

	var empty ReqParam_Resource
	if err := json.Unmarshal([]byte(`{"resource":{},"dramaSeries":[]}`), &empty); err != nil {
		t.Fatalf("unmarshal empty dramaSeries: %v", err)
	}
	if empty.DramaSeries == nil || len(empty.DramaSeries) != 0 {
		t.Fatalf("explicit empty dramaSeries should be non-nil and empty: %#v", empty.DramaSeries)
	}
}
