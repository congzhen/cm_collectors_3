package cmscraper

import "errors"

var (
	ErrInvalidImageURL        = errors.New("无效的图片URL")          // 图片URL格式不正确或无法解析
	ErrHomepageAccessFailed   = errors.New("访问主页失败")            // 无法成功访问网站主页
	ErrImagePageAccessFailed  = errors.New("访问图片页面失败")          // 无法成功访问图片所在页面
	ErrConfigLoadFailed       = errors.New("配置加载失败")            // 配置文件读取或解析失败
	ErrMetadataScrapingFailed = errors.New("元数据刮削失败")           // 从网页抓取元数据过程中发生错误
	ErrNoMetadataFound        = errors.New("无法从任何站点获取元数据")      // 所有配置的站点都无法获取到元数据
	ErrInvalidSelectorConfig  = errors.New("无效的SelectorConfig") // CSS选择器配置不合法
	ErrSearchScrapingFailed   = errors.New("搜索刮削失败")            // 在搜索页面抓取信息失败
	ErrrSearchIDSame          = errors.New("因搜索结果中获取的ID与原始ID相同,停止刮削")
)
