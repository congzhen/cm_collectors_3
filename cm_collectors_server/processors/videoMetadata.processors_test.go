package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"testing"
)

func TestValidateVideoMetadataScope(t *testing.T) {
	if err := validateVideoMetadataScope(models.VideoMetadataScopeAll, nil); err != nil {
		t.Fatalf("all scope should not require ids: %v", err)
	}
	if err := validateVideoMetadataScope(models.VideoMetadataScopeSelected, nil); err == nil {
		t.Fatal("selected scope should require ids")
	}
	if err := validateVideoMetadataScope(models.VideoMetadataScopeSelected, []string{"a", "a"}); err != nil {
		t.Fatalf("selected scope should accept unique non-empty ids: %v", err)
	}
}

func TestValidateCronVideoMetadataScope(t *testing.T) {
	if err := validateCronJobScopeSelection(
		models.VideoMetadataScopeAll, nil, datatype.E_cronJobsType_VideoMetadata,
	); err != nil {
		t.Fatalf("video metadata cron should support all scope: %v", err)
	}
	if err := validateCronJobScopeSelection(
		models.VideoMetadataScopeSelected, []string{"a", "b"}, datatype.E_cronJobsType_VideoMetadata,
	); err != nil {
		t.Fatalf("video metadata cron should support multiple libraries: %v", err)
	}
	if err := validateCronJobScopeSelection(
		models.VideoMetadataScopeSelected, []string{"a", "b"}, datatype.E_cronJobsType_Import,
	); err == nil {
		t.Fatal("legacy cron job should still require exactly one library")
	}
}
