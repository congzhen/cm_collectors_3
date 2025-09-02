package utils

import "time"

// Rand_Intn 返回一个介于0到max之间的随机整数
func Rand_Intn(max int) int {
	// 使用当前时间的纳秒数作为种子来生成伪随机数
	// 通过取模运算确保结果在0到max范围内
	return int(time.Now().UnixNano()) % max
}
