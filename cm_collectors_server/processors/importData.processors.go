package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/utils"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type ImportData struct {
}

func (ImportData) UpdateScanDiskConfig(filesBasesId, defaultConfigJson string) error {
	settingModel := models.FilesBasesSetting{
		ScanDiskJsonData: defaultConfigJson,
	}
	return settingModel.Update(core.DBS(), filesBasesId, &settingModel, []string{"scan_disk_json_data"})
}

// ScanDiskImportPaths 扫描磁盘路径中的视频文件，并更新扫描配置
//
// 该函数会扫描指定路径下的所有视频文件，将扫描配置保存到数据库，
// 并过滤出尚未在数据库中记录的新文件路径。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联扫描配置和过滤已存在的文件
//   - config: 磁盘扫描配置，包含扫描路径和视频文件后缀等信息
//   - saveConfig: 是否保存配置
//
// 返回值:
//   - []string: 不存在于数据库中的新文件路径列表
//   - error: 执行过程中可能出现的错误
func (ImportData) ScanDiskImportPaths(filesBasesId string, config datatype.Config_ScanDisk, saveConfig bool) ([]string, error) {
	filesPaths, err := utils.GetFilesByExtensions(config.ScanDiskPaths, config.VideoSuffixName, true)
	if err != nil {
		return nil, err
	}
	filesPaths = utils.SortFilesByOrder(filesPaths, utils.FileNameAsc)
	db := core.DBS()

	if saveConfig {
		jsonBytes, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		// 转换为字符串
		configJsonString := string(jsonBytes)
		settingModel := models.FilesBasesSetting{
			ScanDiskJsonData: configJsonString,
		}
		err = settingModel.Update(db, filesBasesId, &settingModel, []string{"scan_disk_json_data"})
		if err != nil {
			return nil, err
		}
	}

	if config.ImportMode == datatype.ImportMode_cover {
		return filesPaths, nil
	}
	nonExistingSrcPaths, err := models.ResourcesDramaSeries{}.FilterNonExistingSrcPaths(db, filesBasesId, filesPaths)
	if err != nil {
		return nil, err
	}

	return nonExistingSrcPaths, nil
}

// ScanDiskImportData 扫描磁盘导入数据，创建资源记录
//
// 该函数会检查指定路径的视频文件是否存在，查找相似的图片作为海报，
// 如果未找到图片且配置允许自动创建，则从视频中提取关键帧作为海报，
// 最后创建资源记录。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联创建的资源
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含封面海报后缀名和自动创建海报等配置项
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (t ImportData) ScanDiskImportData(filesBasesId, filePath string, config datatype.Config_ScanDisk) error {
	exists := utils.FileExists(filePath)
	if !exists {
		return fmt.Errorf("文件不存在")
	}
	fileDir := utils.GetDirPathFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)

	// 合并为剧集的入口：
	// - FolderToSeries：老逻辑，认为同一文件夹下的视频都属于同一个资源。
	// - SimilarNameToSeries：新逻辑，只在同一文件夹下继续比较文件名相似度，避免把同目录的不同作品误合并。
	if config.FolderToSeries || config.SimilarNameToSeries {
		resourcesDramaSeries, err := ResourcesDramaSeries{}.FindDramaSeriesSlcBySearchPath(filesBasesId, fileDir)
		if err == nil && config.SimilarNameToSeries && !config.FolderToSeries {
			// “同一文件夹合并”优先级更高；只有未开启完整文件夹合并时，才做相似名称过滤。
			resourcesDramaSeries = t.filterSimilarNameDramaSeries(resourcesDramaSeries, fileName)
		}
		if err == nil && len(*resourcesDramaSeries) > 0 {
			existsDramaSeriesFilePath := false
			//判断filepath是否已在改资源的剧集中，则不处理
			for _, resourcesDramaSeries := range *resourcesDramaSeries {
				if resourcesDramaSeries.Src == filePath {
					existsDramaSeriesFilePath = true
					break
				}
			}
			if !existsDramaSeriesFilePath {
				resourcesID := (*resourcesDramaSeries)[0].ResourcesID
				// 直接写入剧集信息
				err := ResourcesDramaSeries{}.Create(core.DBS(), resourcesID, filePath, len(*resourcesDramaSeries))
				if err != nil {
					return err
				}
				// 是否按名称重新排序剧集
				if config.FolderToSeriesSort {
					err := ResourcesDramaSeries{}.SortBySrc(resourcesID)
					if err != nil {
						return err
					}
				}
			}
			//如果是追加操作，这里已完成
			if config.ImportMode == datatype.ImportMode_append {
				return nil
			}
		}
	}

	coverPosterBase64, coverPosterWidth, coverPosterHeight, err := ImportData{}.GetCoverPosterBase64(filePath, config)

	// 资源标题
	resourceTitle := t.GetResourceTitle(filePath, config)

	resourceDataParam := datatype.ReqParam_Resource{
		Resource: datatype.ReqParam_ResourceBase{
			FilesBasesID:      filesBasesId,
			Title:             resourceTitle,
			Mode:              datatype.E_resourceMode_Movies,
			CoverPosterMode:   config.CoverPosterType,
			CoverPosterWidth:  coverPosterWidth,
			CoverPosterHeight: coverPosterHeight,
		},
		PhotoBase64: coverPosterBase64,
		DramaSeries: []datatype.ReqParam_resourceDramaSeries_Base{
			{Src: filePath},
		},
	}

	if config.AutoCreatePoster {
		resourceDataParam.Resource.Definition = t.VideoDefinition(filePath, config)
	}

	nfoPath := path.Join(fileDir, fileName+".nfo")
	// 如果nfo文件不存在，则从文件夹下所有nfo文件中查找
	if !utils.FileExists(nfoPath) && config.EnableNfoFuzzyMatch {
		//读取文件夹下所有nfo文件
		nfos, err := utils.GetFilesByExtensions([]string{fileDir}, []string{"nfo"}, false)
		if err != nil {
			return err
		}
		// 遍历所有nfo文件，判断文件名是否包含fileName
		for _, nfo := range nfos {
			tmpFileName := utils.GetFileNameFromPath(nfo, false)
			if strings.Contains(tmpFileName, fileName) {
				nfoPath = path.Join(fileDir, tmpFileName+".nfo")
				break
			}
		}
	}
	// 如果nfo文件不存在，则使用随机nfo文件
	if !utils.FileExists(nfoPath) && config.UseRandomNfoIfNoneMatch {
		nfos, err := utils.GetFilesByExtensions([]string{fileDir}, []string{"nfo"}, false)
		if err != nil {
			return err
		}
		if len(nfos) > 0 {
			nfoPath = nfos[0]
		}
	}

	t.Nfo(filesBasesId, nfoPath, config.Nfo, &resourceDataParam)

	//根据filePath地址，查询资源是否已经存在
	resourcesDramaSeriesScl, err := ResourcesDramaSeries{}.ListBySrc(filePath)
	if err != nil {
		return err
	}
	if len(*resourcesDramaSeriesScl) > 0 {
		//如果已经存在，则更新
		resourceID := (*resourcesDramaSeriesScl)[0].ResourcesID
		resourceDataParam.Resource.ID = resourceID
		_, err := Resources{}.UpdateResource(&resourceDataParam, false)
		return err
	} else {
		// 创建资源
		_, err = Resources{}.CreateResource(&resourceDataParam)
		return err
	}
}

// GetResourceTitle 根据配置的命名模式生成资源标题
//
// 该函数根据不同的资源命名模式配置，从文件路径中提取相应的部分作为资源标题。
// 支持目录名、文件名、目录+文件名以及完整路径四种命名方式。
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含资源命名模式等配置项
//
// 返回值:
//   - string: 根据配置生成的资源标题
func (ImportData) GetResourceTitle(filePath string, config datatype.Config_ScanDisk) string {
	// 资源标题
	dirName := utils.GetDirNameFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)
	resourceTitle := fileName
	switch config.ResourceNamingMode {
	case datatype.ResourceNamingModeDirName:
		resourceTitle = dirName
	case datatype.ResourceNamingModeDirFileName:
		resourceTitle = dirName + fileName
	case datatype.ResourceNamingModeFullPathName:
		resourceTitle = filePath
	}
	return resourceTitle
}

// filterSimilarNameDramaSeries 从同一文件夹下已有的剧集资源中筛选“应该归入同一资源”的那一组。
//
// 这里传入的数据可能包含同目录下多个资源的剧集，因此先按 ResourcesID 分组，再判断当前文件名
// 是否与某个资源内的任意剧集文件名/资源标题相似；命中后返回该资源的剧集列表，后续追加剧集时
// 会使用这个列表的第一个 ResourcesID。
//
// 匹配规则按保守到宽松分三层：
//  1. 集数前缀 + 相同主体：如 01xxxxxxx / 02xxxxxxx、01.xxxxxxx / 02.xxxxxxx。
//  2. 标题归一化后相同或互相包含：如 My Show 第01集 / My Show 第02集。
//  3. 归一化标题相似度达到阈值：处理少量字符差异，但避免短标题被误合并。
func (ImportData) filterSimilarNameDramaSeries(resourcesDramaSeries *[]models.DramaSeriesWithResource, fileName string) *[]models.DramaSeriesWithResource {
	if resourcesDramaSeries == nil || len(*resourcesDramaSeries) == 0 {
		return &[]models.DramaSeriesWithResource{}
	}
	targetName := normalizeSeriesName(fileName)
	if len([]rune(targetName)) < 2 {
		// 纯集数型文件名（01、02、EP03、第4集）归一化后会变空。
		// 这种情况不能直接放弃，否则最常见的分集命名无法合并；但也必须限制在已有资源同样是纯集数型时才合并。
		if isSimpleSeriesPartName(fileName) {
			return filterSimpleSeriesPartDramaSeries(resourcesDramaSeries)
		}
		return &[]models.DramaSeriesWithResource{}
	}

	grouped := map[string][]models.DramaSeriesWithResource{}
	groupOrder := []string{}
	groupMatched := map[string]bool{}

	for _, item := range *resourcesDramaSeries {
		if _, ok := grouped[item.ResourcesID]; !ok {
			groupOrder = append(groupOrder, item.ResourcesID)
		}
		grouped[item.ResourcesID] = append(grouped[item.ResourcesID], item)

		// 同时比较剧集文件名和资源标题：
		// - 文件名用于处理已经按文件名导入的资源。
		// - 标题用于处理首次导入后资源标题被 NFO 或命名配置改写的资源。
		srcName := utils.GetFileNameFromPath(item.Src, false)
		if isSimilarSeriesFileName(fileName, srcName) || isSimilarSeriesFileName(fileName, item.Title) || isSimilarSeriesName(targetName, srcName) || isSimilarSeriesName(targetName, item.Title) {
			groupMatched[item.ResourcesID] = true
		}
	}

	for _, resourcesID := range groupOrder {
		if groupMatched[resourcesID] {
			dataList := grouped[resourcesID]
			return &dataList
		}
	}
	return &[]models.DramaSeriesWithResource{}
}

// filterSimpleSeriesPartDramaSeries 专门处理 01、02、03 这种“只有集数没有标题主体”的文件名。
//
// 这类文件名无法通过相似标题判断，因为去掉集数后没有可比较内容。因此只有当同目录已有资源的
// 文件名或标题也属于纯集数型时，才认为它们属于同一连续剧，降低误把普通短标题合并的风险。
func filterSimpleSeriesPartDramaSeries(resourcesDramaSeries *[]models.DramaSeriesWithResource) *[]models.DramaSeriesWithResource {
	grouped := map[string][]models.DramaSeriesWithResource{}
	groupOrder := []string{}
	groupMatched := map[string]bool{}

	for _, item := range *resourcesDramaSeries {
		if _, ok := grouped[item.ResourcesID]; !ok {
			groupOrder = append(groupOrder, item.ResourcesID)
		}
		grouped[item.ResourcesID] = append(grouped[item.ResourcesID], item)

		srcName := utils.GetFileNameFromPath(item.Src, false)
		if isSimpleSeriesPartName(srcName) || isSimpleSeriesPartName(item.Title) {
			groupMatched[item.ResourcesID] = true
		}
	}

	for _, resourcesID := range groupOrder {
		if groupMatched[resourcesID] {
			dataList := grouped[resourcesID]
			return &dataList
		}
	}
	return &[]models.DramaSeriesWithResource{}
}

// isSimilarSeriesName 比较“去掉集数、分段号、括号信息后的标题主体”是否相似。
//
// normalizedTargetName 必须是已经通过 normalizeSeriesName 处理后的目标文件名；
// candidateName 可以是原始文件名或资源标题，本函数会自行归一化。
func isSimilarSeriesName(normalizedTargetName string, candidateName string) bool {
	candidate := normalizeSeriesName(candidateName)
	if len([]rune(candidate)) < 2 {
		return false
	}
	if normalizedTargetName == candidate {
		return true
	}
	// 对较长标题允许包含关系，比如 “show name” 与 “show name special”。
	// 短标题不走包含匹配，避免 “ab” 误命中大量名称。
	if len([]rune(normalizedTargetName)) >= 4 && len([]rune(candidate)) >= 4 {
		if strings.Contains(normalizedTargetName, candidate) || strings.Contains(candidate, normalizedTargetName) {
			return true
		}
	}
	// 0.86 是偏保守的阈值：允许少量字符、空格、符号差异，但不会把主体明显不同的标题并在一起。
	return stringSimilarity(normalizedTargetName, candidate) >= 0.86
}

// isSimilarSeriesFileName 判断“文件名前缀是集数，后面主体相同”的场景。
//
// 典型例子：
// - 01xxxxxxx / 02xxxxxxx
// - 01.xxxxxxx / 02.xxxxxxx
// - 01-短剧 / 02.短剧
//
// 这种命名如果只做整体相似度，短标题时可能分数不稳定；先拆掉开头集数再比较主体更可靠。
func isSimilarSeriesFileName(targetName string, candidateName string) bool {
	targetSuffix, ok := leadingSeriesPartSuffix(targetName)
	if !ok {
		return false
	}
	candidateSuffix, ok := leadingSeriesPartSuffix(candidateName)
	if !ok {
		return false
	}
	return targetSuffix == candidateSuffix || stringSimilarity(targetSuffix, candidateSuffix) >= 0.9
}

// isSimpleSeriesPartName 判断文件名是否只有集数/分段信息，没有可比较的标题主体。
//
// 支持：
// - 01、1、001
// - EP01、Episode 02、Part-03、CD1
// - 第01集、第2话
func isSimpleSeriesPartName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	// 去掉清晰度、来源等括号标记后再判断，避免 01[1080p] 这种命名失效。
	name = regexp.MustCompile(`[（(【\[].*?[）)】\]]`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	patterns := []string{
		`^[0-9]{1,3}$`,
		`^(?i)(e|ep|episode|part|pt|cd|disc|vol)\s*[-_. ]*\d{1,4}$`,
		`^第\s*[0-9一二三四五六七八九十百千万]{1,4}\s*[集话話回期部]$`,
	}
	for _, pattern := range patterns {
		if regexp.MustCompile(pattern).MatchString(name) {
			return true
		}
	}
	return false
}

// leadingSeriesPartSuffix 从“集数前缀 + 标题主体”中提取标题主体。
//
// 例如：
// - 01xxxxxxx  -> xxxxxxx
// - 01.xxxxxxx -> xxxxxxx
// - 02-短剧    -> 短剧
//
// 如果提取出来的主体过短或仍然只是集数，则返回 false，避免 01.02 这种无标题内容被误合并。
func leadingSeriesPartSuffix(name string) (string, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[（(【\[].*?[）)】\]]`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	match := regexp.MustCompile(`^[0-9]{1,3}[\s._\-–—]*(.+)$`).FindStringSubmatch(name)
	if len(match) != 2 {
		return "", false
	}
	suffix := normalizeSeriesName(match[1])
	if len([]rune(suffix)) < 2 || isSimpleSeriesPartName(suffix) {
		return "", false
	}
	return suffix, true
}

// normalizeSeriesName 将文件名/标题归一化为适合相似度比较的“主体名称”。
//
// 处理内容包括：
// - 去掉 S01E02、EP01、Part-03、第12集、尾部纯数字等集数/分段标记。
// - 去掉括号中的清晰度、字幕组、来源等附加信息。
// - 把符号统一压缩为空格，只保留字母、数字和中文。
//
// 注意：对 01、02 这种纯集数名，归一化结果会为空；调用方需要使用 isSimpleSeriesPartName 兜底。
func normalizeSeriesName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacers := []struct {
		pattern string
		repl    string
	}{
		{`(?i)s\d{1,2}\s*e\d{1,3}`, " "},
		{`(?i)\b(e|ep|episode|part|pt|cd|disc|vol)\s*[-_. ]*\d{1,4}\b`, " "},
		{`第\s*[0-9一二三四五六七八九十百千万]+\s*[集话話回期部]`, " "},
		{`[（(【\[].*?[）)】\]]`, " "},
		{`[\s._\-–—]+[0-9一二三四五六七八九十百千万]{1,4}$`, " "},
		{`[0-9]{1,4}$`, " "},
	}
	for _, replacer := range replacers {
		name = regexp.MustCompile(replacer.pattern).ReplaceAllString(name, replacer.repl)
	}

	var builder strings.Builder
	lastWasSpace := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			builder.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			builder.WriteRune(' ')
			lastWasSpace = true
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

// stringSimilarity 使用 Levenshtein 编辑距离计算两个字符串的相似度，返回 0~1。
//
// 1 表示完全一致，0 表示完全不相似。这里按 rune 处理，避免中文被按字节拆分。
func stringSimilarity(a, b string) float64 {
	aRunes := []rune(a)
	bRunes := []rune(b)
	maxLen := len(aRunes)
	if len(bRunes) > maxLen {
		maxLen = len(bRunes)
	}
	if maxLen == 0 {
		return 1
	}
	distance := levenshteinDistance(aRunes, bRunes)
	return 1 - float64(distance)/float64(maxLen)
}

// levenshteinDistance 计算两个 rune 切片之间的编辑距离。
//
// 编辑距离越小表示字符串越接近；支持插入、删除、替换三种操作。
func levenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = minInt(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

// VideoDefinition 获取视频文件的清晰度信息
//
// 该函数通过FFmpeg工具获取视频文件的基本信息，包括宽度和高度，
// 然后根据这些尺寸信息确定视频的清晰度等级并返回
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置信息
//
// 返回值:
//   - string: 视频清晰度标识符（如"4K"、"1080P"等），如果获取失败则返回空字符串

func (ImportData) VideoDefinition(filePath string, config datatype.Config_ScanDisk) string {
	videoBasicInfo, err := processorsffmpeg.VideoInfo{}.GetVideoBasicInfo(filePath)
	if err == nil {
		videoDefinition := processorsffmpeg.VideoInfo{}.GetVideoDefinition(videoBasicInfo.Width, videoBasicInfo.Height)
		return string(videoDefinition)
	}
	return ""
}

// GetCoverPosterBase64 获取视频文件的封面海报Base64编码及相关尺寸信息
//
// 该函数首先在视频文件所在目录查找匹配的图片文件作为封面海报，
// 如果未找到且配置允许自动创建，则从视频中提取关键帧作为海报，
// 然后根据配置对海报进行尺寸调整和裁剪，最后转换为Base64编码返回
//
// 参数:
//   - filePath: 视频文件的完整路径
//   - config: 磁盘扫描配置，包含封面海报后缀名、匹配规则和处理选项等配置项
//
// 返回值:
//   - string: 封面海报的Base64编码字符串
//   - int: 封面海报的宽度
//   - int: 封面海报的高度
//   - error: 执行过程中可能出现的错误
func (t ImportData) GetCoverPosterBase64(filePath string, config datatype.Config_ScanDisk) (string, int, int, error) {
	fileDir := utils.GetDirPathFromFilePath(filePath)
	fileName := utils.GetFileNameFromPath(filePath, false)
	// 查找符合后缀名的图片文件
	imagePaths, err := utils.GetFilesByExtensions([]string{fileDir}, config.CoverPosterSuffixName, false)
	if err != nil {
		return "", 0, 0, err
	}
	// 匹配图片文件
	coverPosterPath := t.findCoverPoster(imagePaths, fileName, config.CoverPosterMatchName, config.CoverPosterFuzzyMatch, config.CoverPosterUseRandomImageIfNoMatch)
	var coverPosterBytes []byte
	var coverPosterBase64 string
	var coverPosterWidth int
	var coverPosterHeight int
	// 如果未找到相似图片，则尝试自动创建海报
	if coverPosterPath == "" && config.AutoCreatePoster {
		// 自动创建海报
		coverPosterBytes, err = processorsffmpeg.Thumbnail{}.ExtractThumbnailPoster(filePath)
		if err != nil {
			core.LogErr(err)
		}
	} else {
		coverPosterBytes, _ = utils.ImageToBytes(coverPosterPath)
	}
	// 获取图片尺寸
	coverPosterWidth, coverPosterHeight, err = utils.GetImageDimensionsFromBytes(coverPosterBytes)
	if err != nil {
		coverPosterWidth = 200
		coverPosterHeight = 200
	}
	// 如果是自动创建的海报，宽度大于400，则缩放海报
	if coverPosterPath == "" && config.AutoCreatePoster && coverPosterWidth > 400 {
		resizeImageBytes, resizeImageWidth, resizeImageHeight, err := utils.ResizeImageByMaxWidth(coverPosterBytes, 400)
		if err == nil {
			// 缩放后的海报数据
			coverPosterBytes = resizeImageBytes
			coverPosterWidth = resizeImageWidth
			coverPosterHeight = resizeImageHeight
		}
	}

	// 封面进行封面尺寸适应裁切
	if config.CoverPosterType >= 0 {
		cropImageBytes, err := utils.ResizeAndCropImage(coverPosterBytes, config.CoverPosterWidth, config.CoverPosterHeight)
		if err == nil {
			coverPosterBytes = cropImageBytes
			coverPosterWidth = config.CoverPosterWidth
			coverPosterHeight = config.CoverPosterHeight
		}
	}

	// 转换为Base64
	coverPosterBase64, _ = utils.ImageBytesToBase64(coverPosterBytes)
	return coverPosterBase64, coverPosterWidth, coverPosterHeight, err
}

// findCoverPoster 在给定的图片路径中查找匹配的封面海报
//
// 该函数根据提供的匹配规则在图片路径中查找与目标文件名匹配的封面海报。
// 可以使用预定义的匹配名称或直接使用目标文件名进行匹配，支持模糊匹配和严格匹配两种模式。
//
// 参数:
//   - imagePaths: 图片文件路径数组，用于在其中查找匹配的封面海报
//   - targetFileName: 目标文件名，当coverPosterMatchName为空时使用此名称进行匹配
//   - coverPosterMatchName: 预定义的封面海报匹配名称数组，用于指定匹配规则
//   - fuzzyMatch: 是否启用模糊匹配模式，true表示启用模糊匹配，false表示严格匹配
//   - coverPosterUseRandomImageIfNoMatch: 是否使用随机图片作为封面海报，true表示使用随机图片，false表示使用默认图片
//
// 返回值:
//   - string: 匹配到的图片文件路径，未找到匹配项时返回空字符串
func (ImportData) findCoverPoster(imagePaths []string, targetFileName string, coverPosterMatchName []string, fuzzyMatch bool, coverPosterUseRandomImageIfNoMatch bool) string {
	// 如果 coverPosterMatchName 为空，则使用 targetFileName 进行匹配
	if len(coverPosterMatchName) == 0 {
		// 查找与targetFileName相近的图片文件名
		for _, imagePath := range imagePaths {
			imageName := utils.GetFileNameFromPath(imagePath, false)
			if fuzzyMatch {
				// 模糊匹配：文件名完全匹配或者包含关系
				if imageName == targetFileName || strings.Contains(imageName, targetFileName) || strings.Contains(targetFileName, imageName) {
					return imagePath
				}
			} else {
				// 严格匹配：文件名完全匹配
				if imageName == targetFileName {
					return imagePath
				}
			}
		}
	} else {
		// 使用 coverPosterMatchName 的值做匹配
		for _, _matchName := range coverPosterMatchName {
			for _, imagePath := range imagePaths {
				imageName := utils.GetFileNameFromPath(imagePath, false)
				matchName := string(_matchName)

				//如果以regex:开头，代表使用这则表达式匹配，后面是正则表达式，其中@filename或者@fileName要被替换成文件名targetFileName
				if strings.HasPrefix(matchName, "regex:") {
					regexPattern := strings.TrimPrefix(matchName, "regex:")
					regex := strings.ReplaceAll(regexPattern, "@fileName", targetFileName)
					regex = strings.ReplaceAll(regex, "@filename", targetFileName)
					match, _ := regexp.MatchString(regex, imageName)
					if match {
						return imagePath
					}
				}
				// 如果是fileName，则将matchName替换成targetFileName，用以匹配文件名
				if _matchName == datatype.CoverPosterMatchName_fileName {
					matchName = targetFileName
				}
				if fuzzyMatch {
					// 模糊匹配：文件名完全匹配或者包含关系
					if imageName == matchName || strings.Contains(imageName, matchName) || strings.Contains(matchName, imageName) {
						return imagePath
					}
				} else {
					// 严格匹配：文件名完全匹配
					if imageName == matchName {
						return imagePath
					}
				}
			}
		}
	}
	if coverPosterUseRandomImageIfNoMatch && len(imagePaths) > 0 {
		// 随机取数组中的一个元素
		randIndex := utils.Rand_Intn(len(imagePaths))
		return imagePaths[randIndex]
	}
	return ""
}

// Nfo 从NFO文件中解析元数据并应用到资源数据
//
// 该函数读取指定路径的NFO文件(XML格式)，解析其中的元数据，
// 并根据配置将解析的数据应用到资源数据结构中。
//
// 参数:
//   - filesBasesId: 文件库ID
//   - nfoPath: NFO文件的完整路径
//   - nfoConfig: NFO配置信息，包含是否启用NFO功能及解析规则
//   - data: 资源数据指针，用于存储从NFO文件解析出的元数据
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (t ImportData) Nfo(filesBasesId, nfoPath string, nfoConfig datatype.Config_ScanDisk_Nfo, data *datatype.ReqParam_Resource) error {
	if !nfoConfig.NfoStatus {
		return nil
	}
	// 检查文件是否存在
	if !utils.FileExists(nfoPath) {
		return nil
	}
	// 读取XML文件
	xmlFile, err := os.Open(nfoPath)
	if err != nil {
		return nil
	}
	defer xmlFile.Close()

	// 解析XML
	byteValue, _ := io.ReadAll(xmlFile)

	// 创建XML解码器
	decoder := xml.NewDecoder(strings.NewReader(string(byteValue)))

	// 解析为通用结构
	rootElement, err := utils.XML_parseXMLToMap(decoder)
	if err != nil {
		return err
	}

	return t.NfoExecData(filesBasesId, nfoPath, rootElement, nfoConfig, data)
}

// NfoExecData 从NFO配置中提取XML数据并填充到资源数据结构中
//
// 该函数根据NFO配置中指定的XPath路径从XML数据中提取各种媒体信息，
// 包括标题、番号、发行日期、简介、标签、演员等，并将这些信息填充到资源数据结构中。
//
// 参数:
//   - filesBasesId: 文件库ID，用于关联标签和演员信息
//   - nfoPath: NFO文件的完整路径，用于定位相关资源文件
//   - rootElement: 已解析的XML根元素数据映射
//   - nfoConfig: NFO配置信息，包含各类字段的XPath路径配置
//   - data: 资源数据指针，用于存储从XML中提取的信息
//
// 返回值:
//   - error: 执行过程中可能出现的错误
func (ImportData) NfoExecData(filesBasesId, nfoPath string, rootElement map[string]interface{}, nfoConfig datatype.Config_ScanDisk_Nfo, data *datatype.ReqParam_Resource) error {
	var xmlData map[string]interface{}

	// 根据Roots配置获取根节点
	if len(nfoConfig.Roots) > 0 {
		// 有指定根节点
		for _, rootPath := range nfoConfig.Roots {
			if data := utils.XML_getXMLValueByPath(rootElement, rootPath); data != nil {
				if mappedData, ok := data.(map[string]interface{}); ok {
					xmlData = mappedData
					break
				}
			}
		}
	} else {
		// 没有指定根节点，使用整个文档
		xmlData = rootElement
	}

	// 如果没有找到有效的数据节点，直接返回
	if xmlData == nil {
		return nil
	}

	// 提取标题
	if len(nfoConfig.Titles) > 0 {
		for _, titlePath := range nfoConfig.Titles {
			if value := utils.XML_getXMLValueByPath(xmlData, titlePath); value != "" {
				data.Resource.Title = value.(string)
				break
			}
		}
	}

	// 提取番号
	if len(nfoConfig.IssueNumbers) > 0 {
		for _, issueNumberPath := range nfoConfig.IssueNumbers {
			if value := utils.XML_getXMLValueByPath(xmlData, issueNumberPath); value != "" {
				data.Resource.IssueNumber = value.(string)
				break
			}
		}
	}

	// 提取发行日期
	if len(nfoConfig.IssuingDates) > 0 {
		for _, issuingDatePath := range nfoConfig.IssuingDates {
			if value := utils.XML_getXMLValueByPath(xmlData, issuingDatePath); value != "" {
				data.Resource.IssuingDate = value.(string)
				break
			}
		}
	}

	// 提取评分
	if len(nfoConfig.Score) > 0 {
		for _, scorePath := range nfoConfig.Score {
			if value := utils.XML_getXMLValueByPath(xmlData, scorePath); value != "" {
				// 将字符串转换为浮点数
				scoreStr := value.(string)
				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					// 转换失败，跳过该评分
					continue
				}

				// 规范化评分到0-10区间
				if score >= 0 && score <= 10 {
					// 0-10区间直接使用，保留一位小数
					data.Resource.Score = math.Round(score*10) / 10
				} else if score > 10 && score <= 100 {
					// 10-100区间除以10，保留一位小数
					normalizedScore := score / 10.0
					data.Resource.Score = math.Round(normalizedScore*10) / 10
				} else {
					// 不在有效区间内，设为默认值0
					data.Resource.Score = 0
				}
				break
			}
		}
	}

	// 提取简介
	if len(nfoConfig.Abstracts) > 0 {
		for _, abstractPath := range nfoConfig.Abstracts {
			if value := utils.XML_getXMLValueByPath(xmlData, abstractPath); value != "" {
				data.Resource.Abstract = value.(string)
				break
			}
		}
	}
	// 提取标签
	if len(nfoConfig.Tags) > 0 {
		for _, tagPath := range nfoConfig.Tags {
			if values := utils.XML_getXMLValuesByPath(xmlData, tagPath); len(values) > 0 {
				for _, value := range values {
					tag, err := Tag{}.TagInfoByNameNotFoundCreate(filesBasesId, value)
					if err != nil {
						continue
					}
					data.Tags = append(data.Tags, tag.ID)
				}
				break
			}
		}
	}

	performerPhotoBase64 := ""
	// 提取演员照片
	if len(nfoConfig.PerformerThumbs) > 0 && nfoPath != "" {
		for _, performerThumbPath := range nfoConfig.PerformerThumbs {
			if values := utils.XML_getXMLValuesByPath(xmlData, performerThumbPath); len(values) > 0 {
				for _, value := range values {
					// 获取nfo文件所在目录
					nfoDir := filepath.Dir(nfoPath)
					// 调用我们新创建的函数处理演员图片
					base64Image, err := getPerformerImage(value, nfoDir)
					if err != nil {
						core.LogErr(err)
						// 处理错误，可以选择记录日志或者忽略
						continue
					}
					if base64Image != "" {
						// 将获取到的base64图片数据赋值给performerPhotoBase64
						performerPhotoBase64 = base64Image
						// 找到第一个有效的图片就跳出循环
						break
					}
				}
				// 如果已经获取到图片，就不再处理其他路径
				if performerPhotoBase64 != "" {
					break
				}
			}
		}
	}

	// 提取演员
	if len(nfoConfig.PerformerNames) > 0 {
		for _, performerPath := range nfoConfig.PerformerNames {
			if values := utils.XML_getXMLValuesByPath(xmlData, performerPath); len(values) > 0 {
				for _, value := range values {
					performer, err := Performer{}.PerformerInfoByNameNotFoundCreate(filesBasesId, value, performerPhotoBase64)
					if err != nil {
						continue
					}
					data.Performers = append(data.Performers, performer.ID)
				}
				break
			}
		}
	}

	return nil
}

// CRHandlePerformerImage 处理演员图片，根据value判断是链接还是本地文件路径
// 如果是链接，则调用爬虫获取图片数据；如果是本地路径，则直接读取文件
// 最后将图片数据转换为base64格式
// 参数:
//   - value: 图片的URL或者相对路径
//   - nfoDir: nfo文件所在的目录路径
//
// 返回值:
//   - string: base64编码的图片数据
//   - error: 错误信息
func getPerformerImage(value string, nfoDir string) (string, error) {
	var imageData []byte
	var err error

	// 不是链接，认为是相对路径，拼接完整路径
	fullPath := filepath.Join(nfoDir, value)
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("本地图片文件不存在: %s", fullPath)
	}

	// 读取本地图片文件
	imageData, err = os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取本地图片文件失败: %v", err)
	}

	// 将图片数据转换为base64格式
	base64Data, err := utils.ImageBytesToBase64(imageData)
	if err != nil {
		return "", fmt.Errorf("转换图片为base64失败: %v", err)
	}

	return base64Data, nil
}
