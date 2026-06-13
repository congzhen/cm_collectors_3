package processors

import (
	"cm_collectors_server/models"
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestAiTagRecordStatusCanEnterProcessing(t *testing.T) {
	tests := map[string]bool{
		"":                                 true,
		models.AiTagRecordStatusPending:    true,
		models.AiTagRecordStatusProcessing: false,
		models.AiTagRecordStatusSuccess:    false,
		models.AiTagRecordStatusFailed:     false,
		models.AiTagRecordStatusSkipped:    false,
	}

	for status, expected := range tests {
		if actual := canAiTagRecordEnterProcessing(status); actual != expected {
			t.Fatalf("canAiTagRecordEnterProcessing(%q) = %v, want %v", status, actual, expected)
		}
	}
}

func TestAiTagRecordClaimable(t *testing.T) {
	claimable, err := isAiTagRecordClaimable(nil, gorm.ErrRecordNotFound)
	if err != nil || !claimable {
		t.Fatalf("record not found should be claimable, claimable=%v err=%v", claimable, err)
	}

	dbErr := errors.New("database failed")
	claimable, err = isAiTagRecordClaimable(nil, dbErr)
	if err != dbErr || claimable {
		t.Fatalf("database error should not be claimable, claimable=%v err=%v", claimable, err)
	}

	success := &models.AiTagAnalysisRecord{Status: models.AiTagRecordStatusSuccess}
	claimable, err = isAiTagRecordClaimable(success, nil)
	if err != nil || claimable {
		t.Fatalf("success record should not be claimable, claimable=%v err=%v", claimable, err)
	}

	pending := &models.AiTagAnalysisRecord{Status: models.AiTagRecordStatusPending}
	claimable, err = isAiTagRecordClaimable(pending, nil)
	if err != nil || !claimable {
		t.Fatalf("pending record should be claimable, claimable=%v err=%v", claimable, err)
	}
}
