package cmscraper

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// ScraperChromeDp_DownLoad 图片下载器 - 使用ChromeDP下载图片
type ScraperChromeDp_DownLoad struct {
	Headless               bool    // 是否使用无头模式
	VisitHome              bool    // 是否访问主页
	EnableScrollSimulation bool    // 是否启用滚动模拟，模拟真实用户浏览行为
	ScrollIntervalFactor   float64 // 滚动间隔系数，控制滚动操作的时间间隔
}

// Download_Images_WithChromeDP 使用ChromeDP下载图片
// 参数：ctx - 上下文, pageURL - 页面URL, imageMap - 图片文件名到URL的映射
// 返回：图片文件名到base64数据的映射和可能的错误
func (s ScraperChromeDp_DownLoad) Download_Images_WithChromeDP(ctx context.Context, pageURL string, imageMap map[string]string) (map[string]string, error) {
	// 解析图片URL以构建referer，防止防盗链
	u, err := url.Parse(pageURL)
	if err != nil {
		LogError(err.Error())
		return nil, fmt.Errorf("%w: %v", ErrInvalidImageURL, err)
	}

	// 更精确地构建referer URL
	referer := u.Scheme + "://" + u.Host + "/"

	// 增强浏览器模拟配置，尽可能模拟真实浏览器行为
	allocCtx, cancel := chromedp.NewExecAllocator(ctx,
		// 浏览器基础配置
		chromedp.Flag("headless", s.Headless),        // 启用无头模式
		chromedp.Flag("no-sandbox", true),            // 无沙箱模式
		chromedp.Flag("disable-gpu", false),          // 启用GPU
		chromedp.Flag("disable-dev-shm-usage", true), // 禁用/dev/shm使用
		chromedp.Flag("window-size", "1920,1080"),    // 设置窗口大小

		// 用户代理设置
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("lang", "zh-CN,zh;q=0.9,en;q=0.8"),

		// 反反爬相关配置，避免被检测为自动化程序
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "TranslateUI"),
		chromedp.Flag("disable-ipc-flooding-protection", true),

		// 更多反检测配置
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-background-networking", false),
		chromedp.Flag("disable-default-apps", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("metrics-recording-only", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("enable-blink-features", ""),

		// 减少跨域限制的参数
		chromedp.Flag("disable-web-security", true), // 禁用同源策略
		chromedp.Flag("disable-features", "CrossSiteDocumentBlockingIfIsolating"),
		chromedp.Flag("disable-site-isolation-trials", true), // 禁用站点隔离

		// 允许访问本地文件
		chromedp.Flag("allow-file-access-from-files", true),
		chromedp.Flag("allow-running-insecure-content", true),

		// 禁用安全策略
		chromedp.Flag("disable-strict-mime-type-checking", true),
	)
	//defer cancel()

	// 创建浏览器上下文
	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	chromeCtx, cancel = context.WithTimeout(chromeCtx, 90*time.Second)
	defer cancel()

	// 初始化结果map，用于存储图片文件名到base64数据的映射
	result := make(map[string]string)
	for key := range imageMap {
		result[key] = "" // 初始化所有key的值为空字符串
	}
	if s.VisitHome {
		// 先访问主页，模拟真实用户行为
		homepageURL := referer

		// 构建主页操作列表
		homepageActions := []chromedp.Action{
			// 启用网络功能
			network.Enable(),
			// 访问主页
			chromedp.Navigate(homepageURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
		}

		// 如果启用了滚动模拟，则添加滚动操作
		if s.EnableScrollSimulation {
			// 随机等待一段时间后滚动到页面1/3处
			sleepDuration := time.Duration(float64(rand.Intn(3)+2) * s.ScrollIntervalFactor)
			homepageActions = append(homepageActions,
				chromedp.Sleep(sleepDuration*time.Second),
				chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight/3);`, nil),
			)

			// 随机等待一段时间后滚动到页面2/3处
			sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
			homepageActions = append(homepageActions,
				chromedp.Sleep(sleepDuration*time.Second),
				chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight*2/3);`, nil),
			)

			// 随机等待一段时间后滚动到页面底部
			sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
			homepageActions = append(homepageActions,
				chromedp.Sleep(sleepDuration*time.Second),
				chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
			)

			// 随机等待一段时间后滚动到页面顶部
			sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
			homepageActions = append(homepageActions,
				chromedp.Sleep(sleepDuration*time.Second),
				chromedp.Evaluate(`window.scrollTo(0, 0);`, nil),
			)

			// 最后随机等待一段时间
			sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
			homepageActions = append(homepageActions,
				chromedp.Sleep(sleepDuration*time.Second),
			)
		} else {
			// 如果未启用滚动模拟，至少添加基本的等待时间
			homepageActions = append(homepageActions,
				chromedp.Sleep(time.Duration(rand.Intn(3)+2)*time.Second),
			)
		}

		// 执行主页访问操作
		err = chromedp.Run(chromeCtx, homepageActions...)
		if err != nil {
			LogError(err.Error())
			return nil, fmt.Errorf("%w: %v", ErrHomepageAccessFailed, err)
		}
	}

	// 构建imageMap的JSON字符串表示，用于在浏览器中创建图片元素
	imageMapJSON := "{"
	first := true
	for key, url := range imageMap {
		if !first {
			imageMapJSON += ","
		}
		// 转义引号和特殊字符
		escapedKey := strings.ReplaceAll(key, `"`, `\"`)
		escapedURL := strings.ReplaceAll(url, `"`, `\"`)
		imageMapJSON += fmt.Sprintf(`"%s":"%s"`, escapedKey, escapedURL)
		first = false
	}
	imageMapJSON += "}"

	// 构建插入所有图片的JavaScript代码
	insertImagesJS := `
		(() => {
			const imageMap = ` + imageMapJSON + `;
			for (const [key, url] of Object.entries(imageMap)) {
				const img = document.createElement('img');
				img.id = "cm_scraper_image_" + key;
				img.src = url;
				document.body.appendChild(img);
			}
			return true;
		})()
	`

	// 构建图片页面操作列表
	pageActions := []chromedp.Action{
		// 启用网络功能
		network.Enable(),
		// 导航到图片页面URL
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	}

	// 添加等待时间
	if s.EnableScrollSimulation {
		sleepDuration := time.Duration(float64(rand.Intn(3)+2) * s.ScrollIntervalFactor)
		pageActions = append(pageActions, chromedp.Sleep(sleepDuration*time.Second))
	} else {
		pageActions = append(pageActions, chromedp.Sleep(time.Duration(rand.Intn(3)+2)*time.Second))
	}

	// 插入图片元素到页面中
	pageActions = append(pageActions, chromedp.Evaluate(insertImagesJS, nil))

	// 如果启用了滚动模拟，则添加滚动操作
	if s.EnableScrollSimulation {
		// 随机等待一段时间后滚动到页面1/3处
		sleepDuration := time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
		pageActions = append(pageActions,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight/3);`, nil),
			chromedp.Sleep(sleepDuration*time.Second),
		)

		// 随机等待一段时间后滚动到页面2/3处
		sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
		pageActions = append(pageActions,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight*2/3);`, nil),
			chromedp.Sleep(sleepDuration*time.Second),
		)

		// 随机等待一段时间后滚动到页面底部
		sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
		pageActions = append(pageActions,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
			chromedp.Sleep(sleepDuration*time.Second),
		)

		// 随机等待一段时间后滚动到页面顶部
		sleepDuration = time.Duration(float64(rand.Intn(2)+1) * s.ScrollIntervalFactor)
		pageActions = append(pageActions,
			chromedp.Evaluate(`window.scrollTo(0, 0);`, nil),
			chromedp.Sleep(sleepDuration*time.Second),
		)
	} else if s.EnableScrollSimulation == false {
		// 如果未启用滚动模拟，添加基本等待时间
		pageActions = append(pageActions, chromedp.Sleep(time.Duration(rand.Intn(2)+1)*time.Second))
	}

	// 等待页面加载完成
	pageActions = append(pageActions, chromedp.Evaluate(`
		new Promise((resolve) => {
			if (document.readyState === 'complete') {
				resolve();
			} else {
				window.addEventListener('load', resolve);
			}
		});
	`, nil))

	// 访问图片页面URL并执行所有操作
	err = chromedp.Run(chromeCtx, pageActions...)
	if err != nil {
		LogError(err.Error())
		return nil, fmt.Errorf("%w: %v", ErrImagePageAccessFailed, err)
	}

	// 为每个图片提取base64数据
	for key, _ := range imageMap {
		var base64Data string

		// 执行自定义JavaScript来提取每个图片
		extractImageJS := `
			(() => {
				// 获取我们之前插入的图片元素
				const img = document.getElementById('cm_scraper_image_` + key + `');
				if (!img) {
					return "";
				}
				// 如果图片已经加载完成
				if (img.complete && img.naturalHeight !== 0) {
					try {
						const canvas = document.createElement('canvas');
						canvas.width = img.naturalWidth;
						canvas.height = img.naturalHeight;
						const ctx = canvas.getContext('2d');
						ctx.drawImage(img, 0, 0);
						document.body.appendChild(canvas);
						return canvas.toDataURL('image/jpeg', 0.9);
					} catch(e) {
						// 如果canvas方法失败，返回空字符串
						const h1 = document.createElement('h1');
						h1.textContent = e.message;
						document.body.appendChild(h1);
						return "";
					}
				}
				// 如果图片还未加载完成，返回空字符串
				return "";
			})()
		`

		// 执行JavaScript提取图片数据
		err = chromedp.Run(chromeCtx, chromedp.Evaluate(extractImageJS, &base64Data))
		if err != nil {
			// 如果出错，保持空字符串
			result[key] = ""
		} else {
			result[key] = base64Data
		}
	}
	/*
		fmt.Println("浏览器和开发者工具已打开，您可以查看网络请求和控制台输出")
		fmt.Println("查看完毕后，请在控制台按回车键继续...")
		fmt.Scanln() // 等待用户按键
	*/
	LogInfo("图片数量: %d", len(result))
	return result, nil
}
