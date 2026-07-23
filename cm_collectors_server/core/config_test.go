package core

import (
	"cm_collectors_server/config"
	"testing"
)

func TestMergeWithDefaultsFillsPerformerAvatarLibraryConfig(t *testing.T) {
	userConfig := &config.Config{}
	mergeWithDefaults(getDefaultConfig(), userConfig)

	if userConfig.PerformerAvatarLibrary.CachePath != "./runtime/cache/gfriends/" {
		t.Fatalf("unexpected avatar cache path: %q", userConfig.PerformerAvatarLibrary.CachePath)
	}
	if userConfig.PerformerAvatarLibrary.DefaultStrategy != "recommended" {
		t.Fatalf("unexpected avatar strategy: %q", userConfig.PerformerAvatarLibrary.DefaultStrategy)
	}
}
