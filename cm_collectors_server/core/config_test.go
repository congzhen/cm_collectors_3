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
	if userConfig.PerformerAvatarLibrary.ClearCacheOnStartup {
		t.Fatal("avatar cache startup cleanup should be disabled by default")
	}
}

func TestMergeWithDefaultsPreservesTabletPlayerChoice(t *testing.T) {
	legacy := &config.Config{}
	mergeWithDefaults(getDefaultConfig(), legacy)
	if legacy.General.LargeMobilePlayer != "desktop" || legacy.General.LargeMobileShortSide != 768 || legacy.General.LargeMobileLongSide != 1024 {
		t.Fatal("legacy configuration must retain desktop playback and the original thresholds")
	}
	custom := &config.Config{General: config.General{
		LargeMobilePlayer: "mobile", LargeMobileShortSide: 800, LargeMobileLongSide: 1200,
	}}
	mergeWithDefaults(getDefaultConfig(), custom)
	if custom.General.LargeMobilePlayer != "mobile" || custom.General.LargeMobileShortSide != 800 || custom.General.LargeMobileLongSide != 1200 {
		t.Fatal("explicit tablet playback configuration must survive default merging")
	}
}
