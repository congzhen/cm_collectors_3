package filter

import "github.com/microcosm-cc/bluemonday"

// SafeHTML是一个用于过滤输入HTML的函数，它使用了bluemonday库的UGCPolicy策略来确保输出的HTML是安全的，不会执行任何恶意代码。
// 参数：
//   inputData string - 需要被过滤的HTML字符串。
// 返回值：
//   string - 过滤后的安全HTML字符串。
func SafeHTML(inputData string) string {
	// 创建并使用UGCPolicy策略进行HTML内容的过滤
	policy := bluemonday.UGCPolicy()
	return policy.Sanitize(inputData)
}
