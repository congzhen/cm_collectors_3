package processors

import (
	"cm_collectors_server/datatype"
	"testing"
)

// 不合法的播放器设置必须在写配置、访问数据库之前被拒绝。
func TestRejectInvalidPlayerConfig(t *testing.T) {
	for _, value := range []datatype.App_Config{
		{LargeMobilePlayer: "unknown"},
		{LargeMobileShortSide: -1},
		{LargeMobileShortSide: 1200, LargeMobileLongSide: 800},
		{LargeMobileShortSide: 768, LargeMobileLongSide: 4097},
	} {
		if err := (App{}).SetConfig(datatype.App_SystemConfig{App_Config: value}); err == nil {
			t.Fatalf("expected invalid player config to be rejected: %+v", value)
		}
	}
}
