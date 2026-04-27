package core

// 该文件用于向core注册与调用函数
// 规范：需要有注册，调用，以及函数本体，且函数本体需要放在core的其他文件中，要注意循环调用问题

// 这个文件用于注册服务器关闭函数，供外部调用以触发服务器优雅关闭
var shutdownHandler func()

// RegisterShutdownHandler 注册关闭函数
func RegisterShutdownHandler(fn func()) {
	shutdownHandler = fn
}

// InvokeShutdownHandler 触发关闭函数，返回是否成功触发
func InvokeShutdownHandler() bool {
	if shutdownHandler == nil {
		return false
	}
	shutdownHandler()
	return true
}
