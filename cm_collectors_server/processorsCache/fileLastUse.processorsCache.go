package processorscache

import (
	"fmt"
	"os"
	"time"
)

// 文件句柄释放时间
const fileHandleExpiration time.Duration = 10 * time.Second

// 文件句柄清理时间
const fileHandleCleanupInterval time.Duration = 30 * time.Second

var cacheFileLastUseHandle = NewGenericLastUseCache[*os.File]("file", fileHandleExpiration, fileHandleCleanupInterval)

type CacheFileLastUse struct {
}

// 获取缓存的文件句柄或创建新的
func (CacheFileLastUse) GetFileHandle(src string) (*os.File, error) {
	if value, ok := cacheFileLastUseHandle.Get(src); ok {
		fmt.Println("######################### 文件缓存:", src)
		return value, nil
	}
	// 打开新文件
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	fmt.Println("######################### 读取缓存:", src)
	cacheFileLastUseHandle.Set(src, file)
	return file, nil
}

// Invalidate 关闭并移除指定路径的缓存句柄，供需要替换真实文件的流程使用。
func (CacheFileLastUse) Invalidate(src string) error {
	file, ok := cacheFileLastUseHandle.Take(src)
	if !ok || file == nil {
		return nil
	}
	return file.Close()
}
