package models

import (
	"cm_collectors_server/utils"
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"sort"
	"strings"
)

var metadataExtensionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

func NormalizeMetadataExtensions(value string) (string, error) {
	set := map[string]bool{}
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
	}) {
		ext := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(part)), ".")
		if !metadataExtensionPattern.MatchString(ext) {
			return "", fmt.Errorf("无效的排除后缀：%s，请填写 jpg、pdf 等后缀，不支持通配符或路径", part)
		}
		set[ext] = true
	}
	if len(set) > 100 {
		return "", fmt.Errorf("自定义排除后缀最多100项")
	}
	list := []string{}
	for ext := range set {
		list = append(list, ext)
	}
	sort.Strings(list)
	return strings.Join(list, ","), nil
}

func MetadataRules(db *gorm.DB) (VideoMetadataSetting, error) {
	var setting VideoMetadataSetting
	err := db.Where("id = ?", "default").Take(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return (VideoMetadataSetting{}).Default(), nil
	}
	return setting, err
}

// MetadataExclusionSQL 和采集入口共用同一规则。规则排除不依赖历史布尔标记，撤销后缀即可恢复资格。
// 调用方必须提供 ds、vm 两个别名。只使用已校验的后缀拼接 SQL。
func MetadataExclusionSQL(db *gorm.DB) (string, error) {
	setting, err := MetadataRules(db)
	if err != nil {
		return "", err
	}
	return MetadataExclusionSQLFor(setting)
}

func MetadataExclusionSQLFor(setting VideoMetadataSetting) (string, error) {
	custom, err := NormalizeMetadataExtensions(setting.ExcludedExtensions)
	if err != nil {
		return "", err
	}
	path := "LOWER(TRIM(ds.src))"
	// 仅 URL 的 ? 和 # 是参数分隔符，本地文件名中的 # 不能截断。
	for _, sep := range []string{"?", "#"} {
		path = "(CASE WHEN (LOWER(TRIM(ds.src)) LIKE 'http://%' OR LOWER(TRIM(ds.src)) LIKE 'https://%') AND INSTR(" + path + ", '" + sep + "') > 0 THEN SUBSTR(" + path + ", 1, INSTR(" + path + ", '" + sep + "') - 1) ELSE " + path + " END)"
	}
	match := func(exts []string) string {
		parts := []string{}
		for _, ext := range exts {
			if ext != "" {
				parts = append(parts, "SUBSTR("+path+fmt.Sprintf(", -%d) = '%s'", len(ext), ext))
			}
		}
		if len(parts) == 0 {
			return "0=1"
		}
		return "(" + strings.Join(parts, " OR ") + ")"
	}
	// 使用精确后缀匹配，避免下划线等字符被当作 SQL 通配符。
	customExts := []string{}
	for _, ext := range strings.Split(custom, ",") {
		if ext != "" {
			customExts = append(customExts, "."+ext)
		}
	}
	system := "0=1"
	if setting.AutoExcludeNonVideo == nil || *setting.AutoExcludeNonVideo {
		system = match(utils.ClearlyNonVideoExtensions())
	}
	return "CASE WHEN COALESCE(ds.video_metadata_classification,'') = 'video' THEN '' " +
		"WHEN COALESCE(ds.video_metadata_classification,'') = 'nonvideo' THEN 'manual' " +
		"WHEN COALESCE(ds.video_metadata_classification,'') = '' AND COALESCE(ds.video_metadata_excluded,0) = 1 THEN 'legacy' " +
		"WHEN vm.probe_status = 'manual' THEN '' WHEN " + system + " THEN 'system' WHEN " + match(customExts) + " THEN 'custom' ELSE '' END", nil
}

// MetadataSourceRule 用于分集新增和改路径时的初始标记；人工分类由调用方保留。
func MetadataSourceRule(setting VideoMetadataSetting, src string) string {
	if (setting.AutoExcludeNonVideo == nil || *setting.AutoExcludeNonVideo) && utils.IsClearlyNonVideoSource(src) {
		return "system"
	}
	ext := strings.TrimPrefix(utils.MediaSourceExtension(src), ".")
	for _, excluded := range strings.Split(setting.ExcludedExtensions, ",") {
		if ext != "" && ext == excluded {
			return "custom"
		}
	}
	return ""
}
