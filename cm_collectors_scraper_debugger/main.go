package main

import (
	cmscraper "cm_collectors_scraper_debugger/api/cm_scraper"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type TestJson struct {
	Headless               bool   `json:"headless"`               // 是否使用无头浏览器
	VisitHome              bool   `json:"visitHome"`              // 是否访问主页
	ScraperConfig          string `json:"scraperConfig"`          // 配置文件路径
	ID                     string `json:"id"`                     // 刮削ID
	Timeout                int    `json:"timeOut"`                // 刮削超时时间
	RetryCount             int    `json:"retryCount"`             // 刮削失败重试次数
	ImageUseTagName        bool   `json:"imageUseTagName"`        // 是否使用标签名作为图片名称
	EnableScrollSimulation bool   `json:"enableScrollSimulation"` // 是否启用滚动模拟
	SavePath               string `json:"savePath"`               // 保存路径
	Log                    bool   `json:"log"`                    // 是否启用日志
	Batch                  bool   `json:"batch"`                  // 是否批量刮削
	BatchSkipExitsNfo      bool   `json:"batchSkipExitsNfo"`      // 是否跳过已存在的nfo文件
	BatchFolderPath        string `json:"batchFolderPath"`        // 批量处理的文件夹路径
	BatchExtensions        string `json:"batchExtensions"`        // 批量处理的文件扩展名
}

func main() {
	// 读取test.json
	data, err := os.ReadFile("test.json")
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		waitForExit()
		return
	}
	var testJson TestJson
	err = json.Unmarshal(data, &testJson)
	if err != nil {
		fmt.Printf("JSON解析失败: %v\n", err)
		waitForExit()
		return
	}

	// 直接执行测试函数，而不是在goroutine中执行
	runCmscraper(testJson)

	waitForExit()
}

func waitForExit() {
	// 提示用户按任意键退出
	fmt.Println("\n按任意键退出...")
	// 等待用户输入
	b := make([]byte, 1)
	os.Stdin.Read(b)
}

func runCmscraper(testJson TestJson) {
	// 加载配置
	config, err := cmscraper.LoadConfig(testJson.ScraperConfig)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	var dataMap map[string]string = make(map[string]string)

	if testJson.Batch && testJson.BatchFolderPath != "" {
		batchExtensionsSlc := strings.Split(testJson.BatchExtensions, ",")
		fileSlc, err := getFilesByExtensions([]string{testJson.BatchFolderPath}, batchExtensionsSlc, true)
		if err != nil {
			fmt.Printf("获取文件列表失败: %v\n", err)
			return
		}
		for _, filePath := range fileSlc {
			id := cmscraper.ParseID(filePath, config)
			//获取文件所在的文件夹路径
			saveFolder := filepath.Dir(filePath)
			if testJson.BatchSkipExitsNfo {
				nfoPath := path.Join(saveFolder, id+".nfo")
				if fileExists(nfoPath) {
					fmt.Println("跳过已存在nfo文件：", nfoPath)
					continue
				}
			}
			dataMap[id] = saveFolder
		}
	} else {
		// 解析文件名获取ID
		filePath := testJson.ID
		id := cmscraper.ParseID(filePath, config)
		saveFolder := fmt.Sprintf("%s@%s", time.Now().Format("20060102150405"), id)
		dataMap[testJson.ID] = filepath.Join(testJson.SavePath, saveFolder)
	}

	if len(dataMap) == 0 {
		log.Fatal("没有找到任何可刮削的数据")
	}
	fmt.Println("等待刮削数据", dataMap)

	for id, saveFolder := range dataMap {
		execCmscraper(config, testJson, id, saveFolder)
	}

}

func execCmscraper(config *cmscraper.ScraperConfig, testJson TestJson, id, saveFolder string) {
	// 创建刮削器
	s := cmscraper.NewScraper(config, testJson.Headless, time.Duration(testJson.Timeout), testJson.RetryCount, true, "scraper.log")
	// 关闭日志
	defer cmscraper.CloseGlobalLogger()

	// 刮削元数据
	ctx := context.Background()
	metadata, pageUrl, err := s.Scrape(ctx, id)
	if err != nil {
		fmt.Printf("刮削失败: %v\n", err)
		return
	}
	for k, v := range *metadata {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("元数据：", metadata)

	//判断元数据是否有效
	if !cmscraper.IsValidMetadata(metadata, config) {
		fmt.Println("元数据无效")
		return
	}

	// 获取图片的base64编码
	images, err := cmscraper.GetMetadataImages(ctx, config, pageUrl, metadata, testJson.ImageUseTagName, testJson.Headless, testJson.VisitHome, testJson.EnableScrollSimulation, 1.0)
	if err != nil {
		cmscraper.LogError("获取图片base64失败: %v", err)
		fmt.Printf("获取图片base64失败: %v\n", err)
	} else {
		for filename, base64Data := range images {
			fmt.Printf("文件名: %s\n", filename)
			// 安全地显示base64数据
			if len(base64Data) > 100 {
				fmt.Printf("Base64: %s...\n", base64Data[:100])
			} else {
				fmt.Printf("Base64: %s\n", base64Data)
			}
			fmt.Printf("Base64长度: %d 字符\n\n", len(base64Data))

			// 保存图片到当前目录
			savePath := filepath.Join(saveFolder, filename)
			err := cmscraper.SaveBase64AsImage(base64Data, savePath, true)
			if err != nil {
				fmt.Printf("保存图片失败: %v\n", err)
			}
		}
	}
	nfo := cmscraper.ToNFO(metadata, &config.Sites[len(config.Sites)-1])
	fmt.Println("NFO内容:")
	fmt.Println(nfo)
	//保存NFO
	nfoFilePath := filepath.Join(saveFolder, fmt.Sprintf("%s.nfo", id))
	err = cmscraper.SaveNfoFile(nfoFilePath, []byte(nfo))
	if err != nil {
		fmt.Printf("保存NFO文件失败: %v\n", err)
	} else {
		fmt.Println("NFO文件保存成功")
	}

	fmt.Println("\n任务执行完成!")
}

// getFilesByExtensions 根据指定的文件扩展名筛选目录中的文件
// 参数:
// dirPaths: 要搜索的目录路径数组，每个路径都是完整的绝对路径
// extensions: 文件扩展名数组，例如 [".mp4", ".avi"] 或 ["mp4", "avi"]
// recursive: 是否递归遍历子目录，true表示递归遍历，false表示只遍历当前目录
// 返回值:
// []string: 匹配的文件绝对路径数组
// error: 错误信息，如果提取成功则为nil
func getFilesByExtensions(dirPaths []string, extensions []string, recursive bool) ([]string, error) {
	var result []string

	if len(dirPaths) == 0 || len(extensions) == 0 {
		return result, nil
	}

	// 标准化扩展名，确保它们以点号开头
	normalizedExtensions := make([]string, len(extensions))
	for i, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			normalizedExtensions[i] = "." + ext
		} else {
			normalizedExtensions[i] = ext
		}
	}

	// 处理每个目录
	for _, dirPath := range dirPaths {
		// 根据是否递归选择不同的遍历方式
		var err error
		if recursive {
			// 递归遍历目录及其子目录
			err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// 跳过目录
				if info.IsDir() {
					return nil
				}

				// 检查文件扩展名是否匹配指定的任何扩展名
				fileExt := strings.ToLower(filepath.Ext(path))
				for _, ext := range normalizedExtensions {
					if fileExt == strings.ToLower(ext) {
						// 将路径转换为标准路径
						path = filepath.ToSlash(path)
						result = append(result, path)
						break
					}
				}

				return nil
			})
		} else {
			// 只遍历当前目录，不递归子目录
			entries, readErr := os.ReadDir(dirPath)
			if readErr != nil {
				err = readErr
			} else {
				for _, entry := range entries {
					// 跳过目录
					if entry.IsDir() {
						continue
					}

					// 检查文件扩展名是否匹配指定的任何扩展名
					fileExt := strings.ToLower(filepath.Ext(entry.Name()))
					for _, ext := range normalizedExtensions {
						if fileExt == strings.ToLower(ext) {
							// 将路径转换为标准路径
							path := filepath.ToSlash(filepath.Join(dirPath, entry.Name()))
							result = append(result, path)
							break
						}
					}
				}
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// getFileNameFromPath 从给定的文件路径中提取文件名
// 参数:
// path: 文件的完整路径
// withExtension: 是否包含扩展名，true表示包含，false表示不包含
// 返回值:
// string: 文件名
func getFileNameFromPath(path string, withExtension bool) string {
	filename := filepath.Base(path)
	if withExtension {
		return filename
	}

	// 移除扩展名
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// fileExists 判断文件或目录是否存在
// 参数:
// path: 要检查的文件或目录路径
// 返回值:
// bool: 文件或目录存在返回true，否则返回false
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
