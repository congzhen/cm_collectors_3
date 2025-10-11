package cmscraper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ScraperConfig 刮削器配置 - 定义整个刮削器的主要配置参数
type ScraperConfig struct {
	Name         string            `json:"name"`                    // 刮削器名称，用于标识不同的刮削器实例
	Search       *SearchConfig     `json:"search"`                  // 搜索配置，用于在目标站点上搜索内容
	Sites        []SiteConfig      `json:"sites"`                   // 站点配置列表，定义多个可抓取的目标站点
	CacheEnabled bool              `json:"cache_enabled"`           // 是否启用缓存机制，避免重复抓取相同内容
	Proxy        *ProxyConfig      `json:"proxy,omitempty"`         // 代理配置，可选配置项
	Headers      map[string]string `json:"headers,omitempty"`       // HTTP请求头，用于模拟不同浏览器请求
	FilePatterns []string          `json:"file_patterns,omitempty"` // 文件名匹配规则，用于从文件名中提取ID
}

// SearchConfig 搜索配置 - 定义如何在站点上进行搜索
type SearchConfig struct {
	URL       string         `json:"url"`       // 搜索URL模板，其中{id}会被实际ID替换
	Selectors SelectorConfig `json:"selectors"` // 选择器配置，定义如何从搜索结果页提取信息
}

// SiteConfig 站点配置 - 定义单个站点的抓取规则
type SiteConfig struct {
	Name           string                    `json:"name"`                      // 站点名称
	URL            string                    `json:"url"`                       // 站点URL模板，其中{id}会被实际ID替换
	Priority       int                       `json:"priority"`                  // 站点优先级，数字越小优先级越高
	Selectors      map[string]SelectorConfig `json:"selectors"`                 // 字段选择器映射，定义如何提取各个字段
	PostProcessors map[string]PostProcessor  `json:"post_processors,omitempty"` // 后处理器映射，对提取的数据进行进一步处理
	Regexps        map[string]string         `json:"regexps"`                   // 正则表达式映射，用于数据清洗
}

// SelectorConfig 定义选择器配置 - 如何从HTML中提取特定数据
type SelectorConfig struct {
	Selector           string         `json:"selector"`            // CSS选择器表达式
	Type               string         `json:"type"`                // 提取类型：text(文本), array(数组), attribute(属性)
	Attribute          string         `json:"attribute"`           // 当Type为attribute时使用，指定要提取的属性名
	FallbackAttributes []string       `json:"fallback_attributes"` // 备用属性列表，主属性不存在时依次尝试
	PostProcess        *PostProcessor `json:"post_process"`        // 后处理配置，对提取的数据进行处理
	JsonSelector       string         `json:"json_selector"`       // JSON选择器
}

// PostProcessor 定义后处理器配置 - 对提取的数据进行二次处理
type PostProcessor struct {
	Type    string `json:"type"`    // 处理类型：regexp(正则), absolute_url(绝对URL), filename(文件名)
	Pattern string `json:"pattern"` // 处理模式，根据Type有不同的含义
}

// ProxyConfig 代理配置 - 定义HTTP代理设置
type ProxyConfig struct {
	Enabled  bool   `json:"enabled"`            // 是否启用代理
	Protocol string `json:"protocol"`           // 协议类型：http, https, socks5
	Host     string `json:"host"`               // 代理服务器地址
	Port     int    `json:"port"`               // 代理服务器端口
	Username string `json:"username,omitempty"` // 代理认证用户名（可选）
	Password string `json:"password,omitempty"` // 代理认证密码（可选）
}

// Scraper 刮削器 - 主要的刮削器结构（注意：实际实现在ScraperChromeDp中）
type Scraper struct {
	config     *ScraperConfig           // 刮削器配置
	cache      map[string]*CachedResult // 缓存结果，以ID为键存储已抓取的数据
	retryDelay time.Duration            // 重试延迟时间
	retryCount int                      // 最大重试次数
}

// CachedResult 缓存结果结构 - 存储缓存的元数据和对应的目标URL
type CachedResult struct {
	Metadata  *map[string]any // 元数据映射
	TargetURL string          // 对应的目标URL
}

// 常见图片扩展名映射表，用于快速判断URL是否指向图片资源
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".bmp": true, ".webp": true, ".tiff": true, ".svg": true,
	".ico": true, ".jfif": true, ".pjpeg": true, ".pjp": true,
}

// 默认文件名解析模式，用于从文件名中提取ID
var defaultFilePatterns = []string{
	`([A-Z0-9]+-[0-9]+)`, // 匹配类似 ABC123-456 的格式
	`([a-z0-9]+-[0-9]+)`, // 匹配类似 abc123-456 的格式
}

// LoadConfig 从文件加载刮削配置
// 参数：filePath - 配置文件路径
// 返回：解析后的配置对象和可能的错误
func LoadConfig(filePath string) (*ScraperConfig, error) {
	// 读取配置文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigLoadFailed, err)
	}

	// 解析JSON配置
	var config ScraperConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigLoadFailed, err)
	}

	// 设置默认值
	if config.CacheEnabled == false {
		config.CacheEnabled = true
	}

	// 如果没有指定文件名解析模式，使用默认模式
	if len(config.FilePatterns) == 0 {
		config.FilePatterns = defaultFilePatterns
	}

	// 为没有设置优先级的站点分配默认优先级
	for i := range config.Sites {
		if config.Sites[i].Priority == 0 {
			config.Sites[i].Priority = i + 1
		}
	}

	return &config, nil
}

// NewScraper 创建一个新的ScraperChromeDp实例
// config: 爬虫配置信息
// headless: 是否启用无头模式
// retryDelay: 重试延迟时间
// retryCount: 重试次数
// logEnabled: 是否启用日志
// logPath: 日志文件路径
// 返回值: ScraperChromeDp实例指针
func NewScraper(config *ScraperConfig, headless bool, retryDelay time.Duration, retryCount int, logEnabled bool, logPath string) *ScraperChromeDp {
	scraper := &ScraperChromeDp{
		headless:   headless,
		config:     config,
		cache:      make(map[string]*CachedResult), // 初始化缓存
		retryDelay: retryDelay,
		retryCount: retryCount,
	}
	if logEnabled {
		// 配置全局日志设置
		SetGlobalLogEnabled(logEnabled)
		if logPath != "" {
			SetGlobalLogFilePath(logPath)
		}
	}
	LogInfo("***创建刮削器***")
	return scraper
}

// ParseID 从文件路径中解析ID
// 参数：filePath - 文件路径, config - 刮削器配置
// 返回：解析出的ID字符串
func ParseID(filePath string, config *ScraperConfig) string {
	filename := filepath.Base(filePath) // 获取文件名
	// 移除扩展名
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)

	// 使用配置文件中的文件名匹配模式
	if config != nil && len(config.FilePatterns) > 0 {
		// 尝试每个模式
		for _, pattern := range config.FilePatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			matches := re.FindStringSubmatch(filename)
			if len(matches) > 1 {
				LogInfo("解析得到的ID: %s", matches[1])
				// 返回第一个捕获组
				return matches[1]
			}
		}
	}

	// 如果没有匹配的模式，返回清理后的文件名
	filename = strings.ReplaceAll(filename, "_", "-")
	filename = strings.ReplaceAll(filename, " ", "-")
	LogInfo("解析得到的ID: %s", filename)
	return filename
}

// GetMetadataImages 从元数据中提取图片URL并下载
// 参数：ctx - 上下文, pageURL - 页面URL, metadata - 元数据,
//
// useTagName - 是否使用标签名作为文件名,
// headless - 是否无头模式,
// visitHome - 是否访问主页,
// enableScrollSimulation - 是否启用滚动模拟,
// scrollIntervalFactor - 滚动间隔系数
//
// 返回：图片文件名到base64数据的映射和可能的错误
func GetMetadataImages(ctx context.Context, pageURL string, metadata *map[string]any, useTagName bool, headless bool, visitHome bool, enableScrollSimulation bool, scrollIntervalFactor float64) (map[string]string, error) {
	images := make(map[string]string) // 存储图片文件名到URL的映射

	// 遍历元数据，提取图片URL
	for key, value := range *metadata {
		switch v := value.(type) {
		case []string:
			// 处理数组类型字段（如多个演员图片）
			for j, item := range v {
				if !isImageURL(item, imageExtensions) {
					continue
				}
				u, err := url.Parse(item)
				if err != nil {
					continue
				}
				var filename string
				if useTagName {
					// 使用标签名+索引作为文件名
					filename = fmt.Sprintf("%s_%d.jpg", key, j)
				} else {
					// 使用URL中的文件名
					filename = filepath.Base(u.Path)
					if filename == "" || filename == "/" {
						filename = fmt.Sprintf("cover_%s.jpg", fmt.Sprintf("%s_%d", key, j))
					}
				}
				images[filename] = item
			}

		case string:
			// 处理字符串类型字段（如封面图片）
			if !isImageURL(v, imageExtensions) {
				continue
			}
			u, err := url.Parse(v)
			if err != nil {
				continue
			}
			var filename string
			if useTagName {
				// 使用标签名作为文件名
				filename = fmt.Sprintf("%s.jpg", key)
				(*metadata)[key] = filename
			} else {
				// 使用URL中的文件名
				filename = filepath.Base(u.Path)
				if filename == "" || filename == "/" {
					filename = fmt.Sprintf("cover_%s.jpg", key)
				}
			}
			images[filename] = v
		}
	}

	// 使用ChromeDP下载图片
	scd := ScraperChromeDp_DownLoad{
		Headless:               headless,
		VisitHome:              visitHome,
		EnableScrollSimulation: enableScrollSimulation,
		ScrollIntervalFactor:   scrollIntervalFactor,
	}
	return scd.Download_Images_WithChromeDP(ctx, pageURL, images)
}

// ToNFO 将元数据转换为NFO格式（媒体库标准格式）
// 参数：metadata - 元数据映射, config - 站点配置
// 返回：NFO格式的XML字符串
func ToNFO(metadata *map[string]any, config *SiteConfig) string {
	if metadata == nil {
		return ""
	}

	// 构建NFO XML头部
	nfo := `<?xml version="1.0" encoding="UTF-8" standalone="yes" ?>
<movie>
`

	// 遍历元数据，生成XML节点
	for key, value := range *metadata {
		switch v := value.(type) {
		case []string:
			// 处理数组类型字段（如演员列表）
			if len(v) > 0 {
				for _, item := range v {
					item = strings.TrimSpace(item)
					if item != "" {
						nfo += fmt.Sprintf("    <%s>%s</%s>\n", key, xmlEscape(item), key)
					}
				}
			}
		case string:
			// 处理字符串类型字段
			if v != "" {
				nfo += fmt.Sprintf("    <%s>%s</%s>\n", key, xmlEscape(v), key)
			}
		}
	}

	nfo += "</movie>"
	return nfo
}

// xmlEscape XML转义 - 将特殊字符转换为XML实体
// 参数：s - 待转义的字符串
// 返回：转义后的字符串
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")   // 转义 &
	s = strings.ReplaceAll(s, "<", "&lt;")    // 转义 <
	s = strings.ReplaceAll(s, ">", "&gt;")    // 转义 >
	s = strings.ReplaceAll(s, "\"", "&quot;") // 转义 "
	s = strings.ReplaceAll(s, "'", "&apos;")  // 转义 '
	return s
}

// IsValidMetadata 检查元数据是否有效，基于配置中定义的选择器
// 参数：metadata - 元数据映射, config - 刮削器配置
// 返回：是否有效的布尔值
func IsValidMetadata(metadata *map[string]any, config *ScraperConfig) bool {
	if metadata == nil || len(*metadata) == 0 {
		LogError("元数据无效")
		return false
	}

	// 收集所有配置中定义的选择器字段
	configFields := make(map[string]bool)

	// 遍历所有站点的 selectors 来确定配置中定义的字段
	for _, site := range config.Sites {
		for fieldName, selectorConfig := range site.Selectors {
			// 只考虑配置了选择器的字段
			if selectorConfig.Selector != "" {
				configFields[fieldName] = true
			}
		}
	}

	// 检查至少有一个配置中定义的字段在元数据中有有效值
	for field := range configFields {
		if value, exists := (*metadata)[field]; exists && !isEmptyValue(value) {
			LogInfo("元数据有效")
			return true
		}
	}

	// 如果没有任何配置的字段有值，则认为元数据无效
	LogError("元数据无效")
	return false
}

// isEmptyValue 检查值是否为空
// 参数：value - 要检查的值
// 返回：是否为空的布尔值
func isEmptyValue(value any) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []string:
		return len(v) == 0 || (len(v) == 1 && strings.TrimSpace(v[0]) == "")
	case []any:
		return len(v) == 0
	case nil:
		return true
	default:
		// 对于其他类型，转换为字符串检查
		str := fmt.Sprintf("%v", v)
		return strings.TrimSpace(str) == ""
	}
}

// applyPostProcess 应用后处理 - 对提取的数据进行二次处理
// 参数：value - 原始值, postProcess - 后处理器配置, baseURL - 基础URL
// 返回：处理后的值
func applyPostProcess(value string, postProcess *PostProcessor, baseURL string) string {
	switch postProcess.Type {
	case "regexp":
		// 正则表达式处理
		if postProcess.Pattern != "" {
			re, err := regexp.Compile(postProcess.Pattern)
			if err != nil {
				// 正则表达式编译失败，返回空字符串
				return ""
			}
			matches := re.FindStringSubmatch(value)
			// 如果匹配成功且有捕获组，则返回第一个捕获组
			if len(matches) > 1 {
				return matches[1]
			}
			// 如果没有匹配或没有捕获组，返回空字符串
			return ""
		}
		return value
	case "absolute_url":
		// 转换为绝对URL
		if value != "" {
			if strings.HasPrefix(value, "//") {
				value = "https:" + value
			} else if strings.HasPrefix(value, "/") {
				parsedBaseURL, _ := url.Parse(baseURL)
				value = parsedBaseURL.Scheme + "://" + parsedBaseURL.Host + value
			}
		}
	case "filename":
		// 提取文件名
		if value != "" {
			filename := filepath.Base(value)
			value = filename
		}
	}
	return value
}

// isImageURL 检查URL是否指向图片资源
// 参数：rawURL - 待检查的URL, imageExtensions - 图片扩展名映射
// 返回：是否为图片URL
func isImageURL(rawURL string, imageExtensions map[string]bool) bool {
	if !strings.HasPrefix(rawURL, "http") {
		return false
	}

	// 检查URL路径中的文件扩展名
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	path := strings.ToLower(u.Path)
	return hasImageExtension(path, imageExtensions)
}

// hasImageExtension 检查路径是否有图片扩展名
// 参数：path - URL路径, imageExtensions - 图片扩展名映射
// 返回：是否有图片扩展名
func hasImageExtension(path string, imageExtensions map[string]bool) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return imageExtensions[ext]
}

// getNestedValue 从嵌套的map中根据点号分隔的路径获取值
// 参数：data - JSON数据映射, path - 点号分隔的路径（如 "pageProps.__N_REDIRECT"）
// 返回：找到的值和是否找到的布尔值
func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	// 按点号分割路径
	keys := strings.Split(path, ".")

	// 从根节点开始遍历
	current := data

	// 遍历除最后一个键之外的所有键
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		// 检查当前层级是否存在该键
		if val, exists := current[key]; exists {
			// 检查值是否为map类型，以便继续深入
			if next, ok := val.(map[string]interface{}); ok {
				current = next
			} else {
				// 如果不是map类型，无法继续深入
				return nil, false
			}
		} else {
			// 键不存在
			return nil, false
		}
	}

	// 获取最后一个键的值
	lastKey := keys[len(keys)-1]
	if val, exists := current[lastKey]; exists {
		return val, true
	}

	return nil, false
}

// SaveBase64AsImage 保存 Base64 字符串为图片
func SaveBase64AsImage(base64Str, filePath string, hasMimeType bool) error {
	// 定义支持的 MIME 类型
	validMimeTypes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
	}
	var replaced string
	if hasMimeType {
		// 查找匹配的 MIME 类型并移除前缀
		for _, mimeType := range validMimeTypes {
			if strings.HasPrefix(base64Str, mimeType) {
				replaced = strings.Replace(base64Str, mimeType, "", 1)
				break
			}
		}
	} else {
		replaced = base64Str

	}

	// 如果没有匹配的 MIME 类型，返回错误
	if replaced == "" {
		err := errors.New("unsupported or invalid base64 string")
		LogError("保存图片 %s 失败: %v", filePath, err)
		return err
	}

	// 解码 Base64 数据
	data, err := base64.StdEncoding.DecodeString(replaced)
	if err != nil {
		LogError("保存图片 %s 失败: %v", filePath, err)
		return err
	}

	// 检查文件夹是否存在，不存在则创建
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			LogError("保存图片 %s 失败: %v", filePath, err)
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 存储为文件
	err = os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		LogError("保存图片 %s 失败: %v", filePath, err)
		return err
	}
	LogInfo("保存图片 %s 成功！", filePath)
	return nil
}

// 保存NFO文件
func SaveNfoFile(nfoFilePath string, data []byte) error {
	err := os.WriteFile(nfoFilePath, []byte(data), 0644)
	if err != nil {
		LogError("保存NFO文件失败:", err)
	}
	LogInfo("保存NFO文件成功！")
	return err
}
