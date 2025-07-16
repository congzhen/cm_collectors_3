package errorMessage

import "fmt"

type ErrorData struct {
	Code int
	Msg  string
}

// Error 实现了error接口的Error方法。
// 该方法返回存储在Msg字段中的错误消息，以便于错误处理和日志记录。
func (e ErrorData) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}
func (e ErrorData) GetCode() int {
	return e.Code
}
func (e ErrorData) GetMsg() string {
	return e.Msg
}

// WithError 返回一个包含错误信息的新 ErrorData 对象
// 该方法通常用于在已有的 ErrorData 对象上添加新的错误信息
// 参数:
//   err error: 需要添加到 ErrorData 对象中的错误
// 返回值:
//   ErrorData: 包含更新后的错误信息的 ErrorData 对象
func (e ErrorData) WithError(err error) ErrorData {
	// 创建并返回一个新的 ErrorData 对象，其 Msg 字段为原始错误信息加上新的错误信息
	return ErrorData{
		Code: e.Code,
		Msg:  e.Msg + " : " + err.Error(),
	}
}

// WrapError 将基础错误和包装错误合并为一个错误。
// 该函数旨在提供更丰富的错误上下文，使调用者更容易理解错误。
// 参数:
//   base: 基础错误，可能是自定义错误类型 ErrorData。
//   wrapErr: 提供额外上下文的包装错误。
// 返回值:
//   返回一个结合了 base 和 wrapErr 的错误。如果 base 是 ErrorData 类型，则调用其 WithError 方法来合并错误；
//   否则使用 fmt.Errorf 来组合错误消息，确保包装错误是 error 类型。
func WrapError(base error, wrapErr error) error {
	// 检查基础错误是否为自定义错误类型 ErrorData
	if e, ok := base.(ErrorData); ok {
		// 如果是，使用自定义错误类型的 WithError 方法来合并错误
		return e.WithError(wrapErr)
	}
	// 如果不是，使用格式字符串方法将基础错误和包装错误组合起来，确保包装错误是 error 类型
	return fmt.Errorf("%v: %w", base, wrapErr)
}

// NewErrorData 创建并返回一个新的ErrorData实例。
// 该函数接收两个参数：code（错误代码）和message（错误消息）。
// 通过将这两个参数封装到ErrorData结构体中，形成一个错误对象。
func NewErrorData(code int, message string) error {
	return ErrorData{
		Code: code,
		Msg:  message,
	}
}

// GetErrorCode 从错误对象中提取错误代码。
// 这个函数尝试将输入的错误对象转换为 ErrorData 类型，如果转换成功，则返回错误代码和 true。
// 如果转换失败，表明输入的错误对象不包含错误代码，此时函数返回 0 和 false。
// 参数:
//   err error: 输入的错误对象。
// 返回值:
//   int: 错误代码，如果无法提取则为 0。
//   bool: 表示是否成功从错误对象中提取了错误代码。
func GetErrorCode(err error) (int, bool) {
	// 尝试将错误对象转换为 ErrorData 类型
	if e, ok := err.(ErrorData); ok {
		// 转换成功，返回错误代码和 true
		return e.GetCode(), true
	}
	// 转换失败，返回 0 和 false
	return 0, false
}

// GetErrorData 尝试将错误接口类型转换为 ErrorData 类型。
// 如果转换成功，则返回转换后的 ErrorData 指针和 true。
// 如果转换失败，则返回 nil 和 false。
func GetErrorData(err error) (*ErrorData, bool) {
	// 使用类型断言尝试将 err 转换为 ErrorData 类型
	if e, ok := err.(ErrorData); ok {
		// 转换成功，返回错误代码和 true
		return &e, true
	}
	// 转换失败，返回 nil 和 false
	return nil, false
}
