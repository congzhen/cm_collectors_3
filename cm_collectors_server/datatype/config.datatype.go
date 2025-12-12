package datatype

type ResourceNamingMode string

const (
	ResourceNamingModeFileName     ResourceNamingMode = "fileName"
	ResourceNamingModeDirName      ResourceNamingMode = "dirName"
	ResourceNamingModeDirFileName  ResourceNamingMode = "dirFileName"
	ResourceNamingModeFullPathName ResourceNamingMode = "fullPathName"
)

type ImportMode string

const (
	ImportMode_append ImportMode = "append"
	ImportMode_cover  ImportMode = "cover"
)

const CoverPosterMatchName_fileName string = "fileName"

// EDetailsDramaSeriesMode 剧集显示模式
type EDetailsDramaSeriesMode string

const (
	FileNameDetailsDramaSeriesMode EDetailsDramaSeriesMode = "fileName"
	DigitDetailsDramaSeriesMode    EDetailsDramaSeriesMode = "digit"
)

// EResourceOpenMode 资源打开方式
type EResourceOpenMode string

const (
	SoftResourceOpenMode       EResourceOpenMode = "soft"
	CloundPlayResourceOpenMode EResourceOpenMode = "cloundPlay"
	SystemResourceOpenMode     EResourceOpenMode = "system"
)

// EResourceOpenModeSoftType 软件内置播放器类型
type EResourceOpenModeSoftType string

const (
	WindowsResourceOpenModeSoftType EResourceOpenModeSoftType = "windows"
	DialogResourceOpenModeSoftType  EResourceOpenModeSoftType = "dialog"
)

// ETagType 标签类型
type ETagType string

const (
	SortTagType       ETagType = "sort"
	CountryTagType    ETagType = "country"
	DefinitionTagType ETagType = "definition"
	YearTagType       ETagType = "year"
	StarTagType       ETagType = "starRating"
	DiyTagType        ETagType = "diyTag"
	PerformerTagType  ETagType = "performer"
	CupTagType        ETagType = "cup"
)

// ICoverPosterData 封面数据
type ICoverPosterData struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   string `json:"type"`
}

// IRouteConversion 路由转换
type IRouteConversion struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Config_Field string

const (
	Config_Field_FilesBases       Config_Field = "config_json_data"
	Config_Field_ScanDisk         Config_Field = "scan_disk_json_data"
	Config_Field_Scraper          Config_Field = "scraper_json_data"
	Config_Field_ScraperPerformer Config_Field = "scraper_performer_json_data"
)

// Config_FilesBases 文件库配置
type Config_FilesBases struct {
	Country                      []string                  `json:"country"`                      // 国家
	CoverDisplayTagAttribute     []string                  `json:"coverDisplayTagAttribute"`     // 封面显示标签属性
	CoverDisplayTag              []string                  `json:"coverDisplayTag"`              // 封面显示标签
	CoverDisplayTagColor         string                    `json:"coverDisplayTagColor"`         // 封面显示标签颜色
	CoverDisplayTagColors        []string                  `json:"coverDisplayTagColors"`        // 封面显示标签颜色
	CoverDisplayTagRgba          string                    `json:"coverDisplayTagRgba"`          // 封面显示标签颜色
	CoverDisplayTagRgbas         []string                  `json:"coverDisplayTagRgbas"`         // 封面显示标签颜色
	CoverDisplayTagFontSize      int                       `json:"coverDisplayTagFontSize"`      // 封面显示标签字体大小
	CoverPosterData              []ICoverPosterData        `json:"coverPosterData"`              // 封面数据
	CoverPosterDataDefaultSelect int                       `json:"coverPosterDataDefaultSelect"` // 封面数据默认选择
	CoverPosterWidthBase         int                       `json:"coverPosterWidthBase"`         // 封面宽度
	CoverPosterWidthStatus       bool                      `json:"coverPosterWidthStatus"`       // 封面宽度状态
	CoverPosterHeightBase        int                       `json:"coverPosterHeightBase"`        // 封面高度
	CoverPosterHeightStatus      bool                      `json:"coverPosterHeightStatus"`      // 封面高度状态
	CoverPosterGap               int                       `json:"coverPosterGap"`               // 封面间隔
	ContentPadding               int                       `json:"contentPadding"`               // 左右空距
	Definition                   []string                  `json:"definition"`                   // 清晰度
	DefinitionFontColor          string                    `json:"definitionFontColor"`          // 清晰度颜色
	DefinitionRgba               string                    `json:"definitionRgba"`               // 清晰度颜色
	DetailsDramaSeriesMode       EDetailsDramaSeriesMode   `json:"detailsDramaSeriesMode"`       // 剧集显示模式
	DirectorText                 string                    `json:"director_Text"`                // 导演显示文字
	HistoryModule                bool                      `json:"historyModule"`                // 历史记录是否开启
	HistoryNumber                int                       `json:"historyNumber"`                // 历史记录数量
	HotModule                    bool                      `json:"hotModule"`                    // 热门资源是否开启
	HotNumber                    int                       `json:"hotNumber"`                    // 热门资源数量
	LeftColumnMode               string                    `json:"leftColumnMode"`               // 左侧栏模式
	LeftColumnWidth              int                       `json:"leftColumnWidth"`              // 左侧栏宽度
	LeftDisplay                  []ETagType                `json:"leftDisplay"`                  // 左侧栏显示内容
	PageLimit                    int                       `json:"pageLimit"`                    // 分页数量
	PerformerPhoto               bool                      `json:"performerPhoto"`               // 左侧栏-演员图片是否开启
	PerformerPreferred           []string                  `json:"performerPreferred"`           // 左侧栏-优先显示演员
	PerformerShowNum             int                       `json:"performerShowNum"`             // 左侧栏-演员显示数量
	PerformerText                string                    `json:"performer_Text"`               // 演员显示名称
	PerformerPhotoDefault        string                    `json:"performer_photo"`              // 默认演员照片
	PlayAtlasImageWidth          int                       `json:"playAtlasImageWidth"`          // 图集-图片宽度
	PlayAtlasMode                string                    `json:"playAtlasMode"`                // 图集-模式
	PlayAtlasPageLimit           int                       `json:"playAtlasPageLimit"`           // 图集-分页数量
	PlayAtlasThumbnail           bool                      `json:"playAtlasThumbnail"`           // 图集-缩略图
	PlayComicMode                string                    `json:"playComicMode"`                // 漫画-模式
	PlayComicrReadingMode        bool                      `json:"playComicrReadingMode"`        // 漫画-阅读模式
	PlugInUnitCup                bool                      `json:"plugInUnit_Cup"`               // 插件单元-cup
	PlugInUnitCupText            string                    `json:"plugInUnit_Cup_Text"`          // 插件单元-cup-文字
	PreviewImageFolder           string                    `json:"previewImageFolder"`           // 预览-图片文件夹
	RandomPosterAutoSize         bool                      `json:"randomPosterAutoSize"`         // 随机海报-自动大小
	RandomPosterWidth            int                       `json:"randomPosterWidth"`            // 随机海报-宽度
	RandomPosterHeight           int                       `json:"randomPosterHeight"`           // 随机海报-高度
	RandomPosterPath             string                    `json:"randomPosterPath"`             // 随机海报-路径
	RandomPosterStatus           bool                      `json:"randomPosterStatus"`           // 随机海报-状态
	ResourceDetailsShowMode      string                    `json:"resourceDetailsShowMode"`      // 资源详情-显示模式
	ResourcesShowMode            string                    `json:"resourcesShowMode"`            // 资源-显示模式
	CoverPosterBoxInfoWidth      int                       `json:"coverPosterBoxInfoWidth"`      // 资源-显示模式-封面海报盒子-信息宽度
	CoverPosterWaterfallColumn   int                       `json:"coverPosterWaterfallColumn"`   // 资源-显示模式-封面海报瀑布流-列数
	CoverImageFit                string                    `json:"coverImageFit"`                // 封面图片填充方式
	CoverTitleAlign              string                    `json:"coverTitleAlign"`              // 封面标题对齐 left right center
	ResourceJustifyContent       string                    `json:"resourceJustifyContent"`       // 资源-显示对齐模式- justify-content   flex-start  center flex-end  space-between  space-around
	RouteConversion              []IRouteConversion        `json:"routeConversion"`              // 路由转换
	ShieldNoPerformerPhoto       bool                      `json:"shieldNoPerformerPhoto"`       // 屏蔽无照片演员
	ShowPreviewImage             bool                      `json:"showPreviewImage"`             // 显示预览图片
	SortMode                     string                    `json:"sortMode"`                     // 排序模式
	TagMode                      string                    `json:"tagMode"`                      // 标签模式
	YouLikeModule                bool                      `json:"youLikeModule"`                // 猜你喜欢-模块
	YouLikeNumber                int                       `json:"youLikeNumber"`                // 猜你喜欢-数量
	YouLikeTagClass              []string                  `json:"youLikeTagClass"`              // 猜你喜欢-标签类
	YouLikeWordNumber            int                       `json:"youLikeWordNumber"`            // 猜你喜欢-匹配词数量
	OpenResModeMovies            EResourceOpenMode         `json:"openResModeMovies"`            // 视频 - 打开方式
	OpenResModeMoviesSoftType    EResourceOpenModeSoftType `json:"openResModeMovies_SoftType"`   // 软件内置播放器类型
	OpenResModeComic             EResourceOpenMode         `json:"openResModeComic"`             // 漫画 - 打开方式
	OpenResModeAtlas             EResourceOpenMode         `json:"openResModeAtlas"`             // 图集 - 打开方式
	SampleStatus                 bool                      `json:"sampleStatus"`                 // 显示剧照
	SampleShowMax                int                       `json:"sampleShowMax"`                // 剧照最大显示数量
	SampleFolder                 string                    `json:"sampleFolder"`                 // 剧照文件夹
}

type Config_ScanDisk struct {
	ScanDiskPaths                      []string            `json:"scanDiskPaths"`
	VideoSuffixName                    []string            `json:"videoSuffixName"`
	AutoGetVideoDefinition             bool                `json:"autoGetVideoDefinition"`
	ResourceNamingMode                 ResourceNamingMode  `json:"resourceNamingMode"`
	ImportMode                         ImportMode          `json:"importMode"`
	CoverPosterMatchName               []string            `json:"coverPosterMatchName"`
	CoverPosterFuzzyMatch              bool                `json:"coverPosterFuzzyMatch"`
	CoverPosterUseRandomImageIfNoMatch bool                `json:"coverPosterUseRandomImageIfNoMatch"`
	CoverPosterSuffixName              []string            `json:"coverPosterSuffixName"`
	CoverPosterType                    int                 `json:"coverPosterType"`
	CoverPosterWidth                   int                 `json:"coverPosterWidth"`
	CoverPosterHeight                  int                 `json:"coverPosterHeight"`
	AutoCreatePoster                   bool                `json:"autoCreatePoster"`
	FolderToSeries                     bool                `json:"folderToSeries"`
	FolderToSeriesSort                 bool                `json:"folderToSeriesSort"`
	EnableNfoFuzzyMatch                bool                `json:"enableNfoFuzzyMatch"`
	UseRandomNfoIfNoneMatch            bool                `json:"useRandomNfoIfNoneMatch"`
	Nfo                                Config_ScanDisk_Nfo `json:"nfo"`
}

type Config_ScanDisk_Nfo struct {
	NfoStatus               bool     `json:"nfoStatus"`
	Roots                   []string `json:"roots"`
	Titles                  []string `json:"titles"`
	IssueNumbers            []string `json:"issueNumbers"`
	IssuingDates            []string `json:"issuingDates"`
	Score                   []string `json:"score"`
	Abstracts               []string `json:"abstracts"`
	Tags                    []string `json:"tags"`
	TagAutoCreate           bool     `json:"tagAutoCreate"`
	PerformerNames          []string `json:"performerNames"`
	PerformerMatchAliasName bool     `json:"performerMatchAliasName"`
	PerformerAutoCreate     bool     `json:"performerAutoCreate"`
	PerformerThumbs         []string `json:"performerThumbs"`
}

type Config_Scraper struct {
	ScanDiskPaths        []string `json:"scanDiskPaths"`
	VideoSuffixName      []string `json:"videoSuffixName"`
	ScraperConfigs       []string `json:"scraperConfigs"`
	RetryCount           int      `json:"retryCount"`
	Timeout              int      `json:"timeout"`
	SkipIfNfoExists      bool     `json:"skipIfNfoExists"`
	SaveNfo              bool     `json:"saveNfo"`
	EnableDownloadImages bool     `json:"enableDownloadImages"`
	UseTagAsImageName    bool     `json:"useTagAsImageName"`
	EnableUserSimulation bool     `json:"enableUserSimulation"`
}
type Config_ScraperPerformer struct {
	ScraperConfig         string           `json:"scraperConfig"`
	Operate               E_ScraperOperate `json:"operate" binding:"required,oneof=update cover"`
	LastScraperUpdateTime string           `json:"lastScraperUpdateTime"`
	Concurrency           int              `json:"concurrency"`
	Timeout               int              `json:"timeout"`
}
