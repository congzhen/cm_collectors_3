package main

import (
	cmscraper "cm_collectors_scraper_debugger/api/cm_scraper"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type TestJson struct {
	Headless               bool   `json:"headless"`
	VisitHome              bool   `json:"visitHome"`
	ScraperConfig          string `json:"scraperConfig"`
	ID                     string `json:"id"`
	Timeout                int    `json:"timeOut"`
	RetryCount             int    `json:"retryCount"`
	ImageUseTagName        bool   `json:"imageUseTagName"`
	EnableScrollSimulation bool   `json:"enableScrollSimulation"`
	SavePath               string `json:"savePath"`
	Log                    bool   `json:"log"`
}

func main() {
	// 读取test.json
	data, err := os.ReadFile("test.json")
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}
	var testJson TestJson
	err = json.Unmarshal(data, &testJson)
	if err != nil {
		log.Fatal("JSON解析失败:", err)
	}
	go test(testJson)
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	fmt.Println("收到退出信号，正在关闭...")
}
func test(testJson TestJson) {
	// 加载配置
	config, err := cmscraper.LoadConfig(testJson.ScraperConfig)
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	// 创建刮削器
	s := cmscraper.NewScraper(config, testJson.Headless, time.Duration(testJson.Timeout), testJson.RetryCount, true, "scraper.log")
	// 关闭日志
	defer cmscraper.CloseGlobalLogger()
	// 解析文件名获取ID
	filePath := testJson.ID
	id := cmscraper.ParseID(filePath, config)

	saveFolder := fmt.Sprintf("%s@%s", time.Now().Format("20060102150405"), id)

	// 刮削元数据
	ctx := context.Background()
	metadata, pageUrl, err := s.Scrape(ctx, id)
	if err != nil {
		log.Fatal("刮削失败:", err)
	}
	for k, v := range *metadata {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("元数据：", metadata)

	//判断元数据是否有效
	if !cmscraper.IsValidMetadata(metadata, config) {
		return
	}

	// 获取图片的base64编码
	images, err := cmscraper.GetMetadataImages(ctx, pageUrl, metadata, testJson.ImageUseTagName, testJson.Headless, testJson.VisitHome, testJson.EnableScrollSimulation, 1.0)
	if err != nil {
		cmscraper.LogError("获取图片base64失败: %v", err)
	} else {
		for filename, base64Data := range images {
			fmt.Println("文件名: %s\n", filename)
			// 安全地显示base64数据
			if len(base64Data) > 100 {
				fmt.Println("Base64: %s...\n", base64Data[:100])
			} else {
				fmt.Println("Base64: %s\n", base64Data)
			}
			fmt.Println("Base64长度: %d 字符\n\n", len(base64Data))

			// 保存图片到当前目录
			savePath := filepath.Join(testJson.SavePath, saveFolder, filename)
			cmscraper.SaveBase64AsImage(base64Data, savePath, true)
		}
	}
	nfo := cmscraper.ToNFO(metadata, &config.Sites[len(config.Sites)-1])
	fmt.Println("NFO内容:")
	fmt.Println(nfo)
	//保存NFO
	nfoFilePath := filepath.Join(testJson.SavePath, saveFolder, fmt.Sprintf("%s.nfo", id))
	cmscraper.SaveNfoFile(nfoFilePath, []byte(nfo))
}
