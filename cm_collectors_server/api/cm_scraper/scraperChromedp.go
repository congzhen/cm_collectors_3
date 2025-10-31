package cmscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// ScraperChromeDp 刮削器 - 使用ChromeDP实现的刮削器
type ScraperChromeDp struct {
	browserPath string                   //浏览器路径
	headless    bool                     // 是否使用无头模式
	config      *ScraperConfig           // 刮削器配置
	cache       map[string]*CachedResult // 缓存结果，以ID为键存储已抓取的数据
	retryDelay  time.Duration            // 重试延迟时间
	retryCount  int                      // 最大重试次数
}

func GetNewExecAllocator(headless bool, proxyConfig *ProxyConfig, browserPath string) (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", headless),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"),

		chromedp.Flag("lang", "zh-CN,zh;q=0.9,en;q=0.8"),
		chromedp.Flag("charset", "utf-8"),
		chromedp.Flag("accept-charset", "utf-8"),

		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "TranslateUI"),
		chromedp.Flag("disable-ipc-flooding-protection", true),

		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-background-networking", false),
		chromedp.Flag("disable-default-apps", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("metrics-recording-only", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("enable-blink-features", ""),

		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "CrossSiteDocumentBlockingIfIsolating"),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),

		chromedp.Flag("allow-file-access-from-files", true),
		chromedp.Flag("allow-running-insecure-content", true),

		chromedp.Flag("disable-strict-mime-type-checking", true),
	}

	// 如果提供了浏览器路径，则使用指定的浏览器
	if browserPath != "" {
		opts = append(opts, chromedp.ExecPath(browserPath))
	}

	// 如果启用了代理配置，则添加代理设置
	if proxyConfig != nil && proxyConfig.Enabled {
		var proxyServer string
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			// 带认证的代理
			proxyServer = fmt.Sprintf("%s://%s:%s@%s:%d",
				proxyConfig.Protocol,
				proxyConfig.Username,
				proxyConfig.Password,
				proxyConfig.Host,
				proxyConfig.Port)
		} else {
			// 不带认证的代理
			proxyServer = fmt.Sprintf("%s://%s:%d",
				proxyConfig.Protocol,
				proxyConfig.Host,
				proxyConfig.Port)
		}
		opts = append(opts, chromedp.Flag("proxy-server", proxyServer))
	}

	return chromedp.NewExecAllocator(context.Background(), opts...)
}

// Scrape 根据ID刮削元数据 - 主要的刮削入口函数
// 参数：ctx - 上下文, id - 要刮削的ID
// 返回：元数据映射、目标URL和可能的错误
func (s *ScraperChromeDp) Scrape(ctx context.Context, id string) (*map[string]any, string, error) {
	// 检查缓存，如果启用缓存且已存在该ID的数据，则直接返回缓存结果
	if s.config.CacheEnabled {
		if cachedResult, exists := s.cache[id]; exists {
			LogInfo("刮削页面:", cachedResult.TargetURL)
			return cachedResult.Metadata, cachedResult.TargetURL, nil
		}
	}

	var err error
	// 如果配置了搜索功能，先执行搜索获取实际ID
	if s.config.Search != nil {
		id, err = s.scrapeSearch(ctx, s.config.Search, id)
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrSearchScrapingFailed, err)
		}
		LogInfo("从搜索结果中获取的ID: %s", id)
	}

	// 按优先级排序站点，优先级数字越小越优先
	sites := make([]SiteConfig, len(s.config.Sites))
	copy(sites, s.config.Sites)
	sort.Slice(sites, func(i, j int) bool {
		return sites[i].Priority < sites[j].Priority
	})

	// 按优先级顺序尝试每个站点
	for _, site := range sites {
		metadata, targetURL, err := s.scrapeSite(ctx, site, id)
		if err == nil && metadata != nil {
			// 添加到缓存
			if s.config.CacheEnabled {
				s.cache[id] = &CachedResult{
					Metadata:  metadata,
					TargetURL: targetURL,
				}
			}
			LogInfo("刮削页面:", targetURL)
			return metadata, targetURL, nil
		}
		// 记录错误但继续尝试下一个站点
		LogError("从站点 %s 刮削失败: %v", site.Name, err)
	}
	LogError(ErrNoMetadataFound.Error())
	return nil, "", ErrNoMetadataFound
}

// scrapeSearch 在搜索页面执行搜索并提取真实ID
// 参数：ctx - 上下文, searchConfig - 搜索配置, id - 搜索关键词
// 返回：搜索到的真实ID和可能的错误
func (s *ScraperChromeDp) scrapeSearch(ctx context.Context, searchConfig *SearchConfig, id string) (string, error) {
	// 创建 Chrome 实例，配置各种参数以模拟真实浏览器
	allocCtx, cancel := GetNewExecAllocator(s.headless, s.config.Proxy, s.browserPath)
	defer cancel()

	// 创建浏览器上下文
	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	chromeCtx, cancel = context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	// 构建目标URL，将{id}替换为实际ID
	targetURL := strings.Replace(searchConfig.URL, "{id}", id, -1)
	LogInfo("正在访问: %s", targetURL)

	// 页面内容变量，用于存储获取到的HTML
	var htmlContent string

	// 访问页面，最多重试3次
	var err error
	for i := 0; i < s.retryCount; i++ {
		err = chromedp.Run(chromeCtx,
			chromedp.Navigate(targetURL),                               // 导航到目标页面
			chromedp.WaitReady("body", chromedp.ByQuery),               // 等待body元素就绪
			chromedp.Sleep(2*time.Second),                              // 等待JavaScript执行
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery), // 获取完整HTML
		)

		if err == nil {
			// 打印获取到的HTML长度以供调试
			LogInfo("成功获取页面内容，HTML长度: %d", len(htmlContent))
			if len(htmlContent) < 100 {
				LogInfo("获取到的HTML内容: %s", htmlContent)
			}
			break
		}
		if s.retryCount > 1 {
			LogError("第 %d 次尝试失败: %v", i+1, err)
			time.Sleep(s.retryDelay * time.Second)
		}

	}

	if err != nil {
		LogError(err.Error())
		return id, fmt.Errorf("%w: %v", ErrSearchScrapingFailed, err)
	}

	// 使用 goquery 解析页面HTML内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		LogError(err.Error())
		return id, fmt.Errorf("%w: %v", ErrSearchScrapingFailed, err)
	}

	// 根据config中的selectors获取数据
	if searchConfig.Selectors.Selector == "" {
		LogError(err.Error())
		return id, fmt.Errorf("%w: %v", ErrInvalidSelectorConfig, searchConfig)
	}

	value := id
	switch searchConfig.Selectors.Type {
	case "html-json":
		// 默认文本类型处理 - 提取元素的文本内容
		jsonStr := ""
		doc.Find(searchConfig.Selectors.Selector).Each(func(i int, sel *goquery.Selection) {
			jsonStr = strings.TrimSpace(sel.Text())
		})
		// 数据返回的是json格式的数据，根据searchConfig.Selectors.Selector（使用.表示层级） 查找json的数据
		// 解析JSON数据
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
			LogError(err.Error())
			return id, fmt.Errorf("JSON解析失败: %v", err)
		}
		// 根据点号分隔的路径获取嵌套值
		if val, exists := getNestedValue(jsonData, searchConfig.Selectors.JsonSelector); exists {
			// 将值转换为字符串
			if strVal, ok := val.(string); ok {
				value = strVal
			} else {
				// 如果不是字符串，转换为字符串
				value = fmt.Sprintf("%v", val)
			}
		} else {
			err := fmt.Errorf("在JSON中未找到路径: %s", searchConfig.Selectors.Selector)
			LogError(err.Error())
			return id, err
		}

	case "attribute":
		// 属性类型处理 - 提取元素的属性值
		doc.Find(searchConfig.Selectors.Selector).Each(func(i int, sel *goquery.Selection) {
			// 首先尝试主属性
			attrValue, exists := sel.Attr(searchConfig.Selectors.Attribute)
			if !exists && len(searchConfig.Selectors.FallbackAttributes) > 0 {
				// 尝试备用属性
				for _, attr := range searchConfig.Selectors.FallbackAttributes {
					attrValue, exists = sel.Attr(attr)
					if exists {
						break
					}
				}
			}
			if exists {
				value = strings.TrimSpace(attrValue)
			}
		})

	default:
		// 默认文本类型处理 - 提取元素的文本内容
		doc.Find(searchConfig.Selectors.Selector).Each(func(i int, sel *goquery.Selection) {
			if value == "" { // 只取第一个匹配项
				value = strings.TrimSpace(sel.Text())
			}
		})
		// 应用后处理
		if searchConfig.Selectors.PostProcess != nil {
			value = applyPostProcess(value, searchConfig.Selectors.PostProcess, targetURL)
		}
	}

	return value, nil
}

// scrapeSite 从特定站点刮削数据
// 参数：ctx - 上下文, site - 站点配置, id - 要刮削的ID
// 返回：元数据映射、目标URL和可能的错误
func (s *ScraperChromeDp) scrapeSite(ctx context.Context, site SiteConfig, id string) (*map[string]any, string, error) {
	// 创建 Chrome 实例，配置各种参数以模拟真实浏览器
	allocCtx, cancel := GetNewExecAllocator(s.headless, s.config.Proxy, s.browserPath)
	defer cancel()

	// 读取config里的headers,并设置请求头
	headers := make(map[string]interface{})
	for k, v := range s.config.Headers {
		headers[k] = v
	}

	// 创建浏览器上下文
	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	chromeCtx, cancel = context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	// 构建目标URL，将{id}替换为实际ID
	targetURL := strings.Replace(site.URL, "{id}", id, -1)
	LogInfo("正在访问: %s", targetURL)

	// 页面内容变量，用于存储获取到的HTML
	var htmlContent string

	// 访问页面，最多重试3次
	var err error
	for i := 0; i < s.retryCount; i++ {
		err = chromedp.Run(chromeCtx,
			chromedp.Navigate(targetURL),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(2*time.Second), // 等待JavaScript执行
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)

		if err == nil {
			// 打印获取到的HTML长度以供调试
			LogInfo("成功获取页面内容，HTML长度: %d", len(htmlContent))
			if len(htmlContent) < 100 {
				LogInfo("获取到的HTML内容: %s", htmlContent)
			}
			break
		}
		if s.retryCount > 1 {
			err := fmt.Errorf("第 %d 次尝试失败: %v\n", i+1, err)
			LogError(err.Error())
			time.Sleep(s.retryDelay * time.Second)
		}

	}

	if err != nil {
		LogError(err.Error())
		return nil, "", fmt.Errorf("%w: %v", ErrMetadataScrapingFailed, err)
	}

	// 使用 goquery 解析页面HTML内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		LogError(err.Error())
		return nil, "", fmt.Errorf("%w: %v", ErrMetadataScrapingFailed, err)
	}

	// 创建元数据对象，用于存储提取的数据
	metadata := map[string]any{}

	// 根据config中的selectors获取数据
	for fieldName, selectorConfig := range site.Selectors {
		if selectorConfig.Selector == "" {
			metadata[fieldName] = ""
			continue
		}

		switch selectorConfig.Type {
		case "array":
			// 数组类型处理 - 用于提取多个值（如演员列表）
			var items []string
			doc.Find(selectorConfig.Selector).Each(func(i int, sel *goquery.Selection) {
				var value string
				var exists bool

				// 如果指定了属性，则提取属性值
				if selectorConfig.Attribute != "" {
					// 首先尝试主属性
					value, exists = sel.Attr(selectorConfig.Attribute)
					if !exists && len(selectorConfig.FallbackAttributes) > 0 {
						// 尝试备用属性
						for _, attr := range selectorConfig.FallbackAttributes {
							value, exists = sel.Attr(attr)
							if exists {
								break
							}
						}
					}

					if exists {
						value = strings.TrimSpace(value)
					}
				} else {
					// 如果没有指定属性，则使用文本内容
					value = strings.TrimSpace(sel.Text())
				}

				// 只有当值不为空时才添加到数组中
				if value != "" {
					// 对数组中的每个元素应用后处理
					if selectorConfig.PostProcess != nil {
						value = applyPostProcess(value, selectorConfig.PostProcess, targetURL)
					}
					items = append(items, value)
				}
			})
			metadata[fieldName] = items // 直接存储数组

		case "attribute":
			// 属性类型处理 - 提取单个元素的属性值
			value := ""
			doc.Find(selectorConfig.Selector).Each(func(i int, sel *goquery.Selection) {
				// 首先尝试主属性
				attrValue, exists := sel.Attr(selectorConfig.Attribute)
				if !exists && len(selectorConfig.FallbackAttributes) > 0 {
					// 尝试备用属性
					for _, attr := range selectorConfig.FallbackAttributes {
						attrValue, exists = sel.Attr(attr)
						if exists {
							break
						}
					}
				}

				if exists {
					value = strings.TrimSpace(attrValue)
				}
			})
			// 应用后处理
			if selectorConfig.PostProcess != nil {
				value = applyPostProcess(value, selectorConfig.PostProcess, targetURL)
			}
			metadata[fieldName] = value

		default:
			// 默认文本类型处理 - 提取单个元素的文本内容
			value := ""
			doc.Find(selectorConfig.Selector).Each(func(i int, sel *goquery.Selection) {
				if value == "" { // 只取第一个匹配项
					value = strings.TrimSpace(sel.Text())
				}
			})
			// 应用后处理
			if selectorConfig.PostProcess != nil {
				value = applyPostProcess(value, selectorConfig.PostProcess, targetURL)
			}
			metadata[fieldName] = value
		}
	}

	// 应用后处理器（正则表达式等）
	for field, processor := range site.PostProcessors {
		if value, exists := metadata[field]; exists {
			if strValue, ok := value.(string); ok && strValue != "" {
				switch processor.Type {
				case "regexp":
					// 正则表达式后处理
					re, err := regexp.Compile(processor.Pattern)
					if err != nil {
						continue
					}
					matches := re.FindStringSubmatch(strValue)
					if len(matches) > 1 {
						metadata[field] = matches[1]
					}
				case "split":
					// 分割字符串后处理
					parts := strings.Split(strValue, processor.Pattern)
					metadata[field] = parts
				}

			}
		}
	}

	return &metadata, targetURL, nil
}
