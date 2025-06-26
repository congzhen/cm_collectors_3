package utils

import "github.com/mozillazg/go-pinyin"

// 获取字符串的拼音首字母
func PinyinInitials(s string) string {
	if s == "" {
		return ""
	}
	var initials string
	args := &pinyin.Args{}
	results := pinyin.LazyConvert(s, args)
	for _, result := range results {
		if len(result) > 0 {
			initials += string(result[0])
		}
	}
	return initials
}
