package models

import (
	"cm_collectors_server/datatype"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SharedLibraryConfig 只存放明确允许共享的参数，不存路径、标签或演员引用。
type SharedLibraryConfig struct {
	Module   string `json:"module" gorm:"primaryKey;size:32"`
	Revision int    `json:"revision"`
	Config   string `json:"-" gorm:"type:text"`
}

func (SharedLibraryConfig) TableName() string { return "shared_library_config" }

type LibraryConfigFollow struct {
	LocalSnapshot *string `json:"-" gorm:"type:text"`
	FilesBasesID  string  `json:"filesBasesId" gorm:"column:filesBases_id;primaryKey;type:char(20)"`
	Module        string  `json:"module" gorm:"primaryKey;size:32"`
}

func (LibraryConfigFollow) TableName() string { return "library_config_follow" }

type SharedConfigState struct {
	CanRestore bool                   `json:"canRestore"`
	Module     string                 `json:"module"`
	Available  bool                   `json:"available"`
	Following  bool                   `json:"following"`
	Revision   int                    `json:"revision"`
	Fields     []string               `json:"fields"`
	Config     map[string]interface{} `json:"config"`
	Libraries  []SharedConfigLibrary  `json:"libraries"`
}

type SharedConfigLibrary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// 白名单也是前端字段锁定的来源。封面预设列表及索引相互依赖，保留本库独立。
func SharedConfigFields(module string) []string {
	if module == "display" {
		return strings.Fields(`country resourceSort definition leftDisplay leftColumnMode leftColumnWidth leftColumnFloatAutoHide
		 tagMode tagFixedModeRowShowNum showCustomTagResourceCount performerPhoto shieldNoPerformerPhoto performerShowNum
		 pageLimit resourcesShowMode showVideoDuration coverPosterBoxInfoWidth coverPosterWaterfallColumn coverTitleAlign
		 resourceJustifyContent detailsDramaSeriesMode resourceDetailsShowMode detailsVisibleFields coverDisplayTagAttribute
		 coverDisplayTagRgbas coverDisplayTagColors coverDisplayTagFontSize casualViewModule casualViewNumber historyModule
		 historyNumber hotModule hotNumber sampleStatus sampleShowMax videoPreviewGeneration videoPreviewInterval videoMomentOpenMode openResModeMovies openResModeMovies_SoftType
		 openResModeComic openResModeAtlas videoPreviewImageCount performer_Text director_Text showPerformerResourceCount
		 plugInUnit_Cup plugInUnit_Cup_Text coverPosterWidthStatus coverPosterWidthBase coverPosterHeightStatus
		 coverPosterHeightBase coverPosterGap contentPadding`)
	}
	if module == "import" {
		return strings.Fields(`videoSuffixName autoGetVideoDefinition resourceNamingMode importMode coverPosterMatchName
          coverPosterFuzzyMatch coverPosterUseRandomImageIfNoMatch coverPosterSuffixName autoCreatePoster folderToSeries
          similarNameToSeries folderToSeriesSortMode enableNfoFuzzyMatch useRandomNfoIfNoneMatch nfo`)
	}
	if module == "scraper" {
		return strings.Fields(`videoSuffixName scraperConfigs concurrency retryCount timeout skipIfNfoExists saveNfo
          enableDownloadImages useTagAsImageName enableUserSimulation`)
	}
	return nil
}

func sharedConfigColumn(module string) string {
	switch module {
	case "display":
		return "config_json_data"
	case "import":
		return "scan_disk_json_data"
	case "scraper":
		return "scraper_json_data"
	}
	return ""
}

func parseConfig(raw string) (map[string]interface{}, error) {
	value := map[string]interface{}{}
	if raw == "" {
		return value, nil
	}
	if err := json.Unmarshal([]byte(raw), &value); err != nil || value == nil {
		return nil, errors.New("配置必须为有效的 JSON 对象")
	}
	return value, nil
}

func sharedFieldsOnly(module string, value map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range SharedConfigFields(module) {
		if v, ok := value[field]; ok {
			result[field] = v
		}
	}
	return result
}

func SharedConfigStatus(db *gorm.DB, id, module string) (*SharedConfigState, error) {
	fields := SharedConfigFields(module)
	if len(fields) == 0 {
		return nil, errors.New("不支持的公共配置分组")
	}
	state := &SharedConfigState{Module: module, Fields: fields, Config: map[string]interface{}{}, Libraries: []SharedConfigLibrary{}}
	var config SharedLibraryConfig
	err := db.Where("module = ?", module).Take(&config).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		state.Available, state.Revision = true, config.Revision
		state.Config, err = parseConfig(config.Config)
		if err != nil {
			return nil, err
		}
	}
	var count int64
	if err := db.Model(&LibraryConfigFollow{}).Where("filesBases_id = ? AND module = ?", id, module).Count(&count).Error; err != nil {
		return nil, err
	}
	state.Following = count > 0
	if state.Following {
		var follow LibraryConfigFollow
		if err := db.Where("filesBases_id = ? AND module = ?", id, module).Take(&follow).Error; err != nil {
			return nil, err
		}
		state.CanRestore = follow.LocalSnapshot != nil
	}
	err = db.Table((LibraryConfigFollow{}).TableName()+" AS f").Select("b.id, b.name").Joins("JOIN filesBases AS b ON b.id = f.filesBases_id").Where("f.module = ?", module).Order("b.sort, b.id").Scan(&state.Libraries).Error
	return state, err
}

// EffectiveLibraryConfig 是普通配置读取的唯一叠加点；独立库原样返回，保留空配置语义。
func EffectiveLibraryConfig(db *gorm.DB, id, module, raw string) (string, error) {
	if sharedConfigColumn(module) == "" {
		return raw, nil
	}
	var follow LibraryConfigFollow
	err := db.Where("filesBases_id = ? AND module = ?", id, module).Take(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return raw, nil
	}
	if err != nil {
		return "", err
	}
	var config SharedLibraryConfig
	if err := db.Where("module = ?", module).Take(&config).Error; err != nil {
		return "", fmt.Errorf("公共配置不可用，请检查配置: %w", err)
	}
	local, err := parseConfig(raw)
	if err != nil {
		return "", err
	}
	shared, err := parseConfig(config.Config)
	if err != nil {
		return "", err
	}
	if module == "import" {
		if _, ok := local["coverPosterType"]; !ok {
			local["coverPosterType"] = -1
		}
	}
	for k, v := range sharedFieldsOnly(module, shared) {
		local[k] = v
	}
	data, err := json.Marshal(local)
	return string(data), err
}

// GuardSharedConfigWrite 防止旧页面、快捷操作及任务保存绕过跟随边界。
func GuardSharedConfigWrite(db *gorm.DB, id, module, incoming string) error {
	var follow LibraryConfigFollow
	err := db.Where("filesBases_id = ? AND module = ?", id, module).Take(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	var shared SharedLibraryConfig
	if err := db.Where("module = ?", module).Take(&shared).Error; err != nil {
		return fmt.Errorf("公共配置不可用: %w", err)
	}
	expectedConfig, err := parseConfig(shared.Config)
	if err != nil {
		return err
	}
	data, err := parseConfig(incoming)
	if err != nil {
		return err
	}
	for k, expected := range expectedConfig {
		if !reflect.DeepEqual(data[k], expected) {
			return errors.New("该分组正在跟随公共配置，请先关闭跟随再修改，或重新加载最新配置")
		}
	}
	return nil
}

func SaveSharedConfig(db *gorm.DB, module string, revision int, raw string) error {
	if len(SharedConfigFields(module)) == 0 {
		return errors.New("不支持的公共配置分组")
	}
	value, err := parseConfig(raw)
	if err != nil {
		return err
	}
	value = sharedFieldsOnly(module, value)
	// 要求完整白名单，避免从残缺旧配置创建出随库默认值漂移的公共配置。
	if len(value) != len(SharedConfigFields(module)) {
		return errors.New("公共配置不完整，请重新打开设置页后再保存")
	}
	for _, v := range value {
		if v == nil {
			return errors.New("公共配置不能包含空值")
		}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	switch module {
	case "display":
		var typed datatype.Config_FilesBases
		if err := json.Unmarshal(data, &typed); err != nil {
			return fmt.Errorf("公共配置参数类型错误: %w", err)
		}
		if typed.VideoPreviewGeneration != "playback" && typed.VideoPreviewGeneration != "manual" {
			return errors.New("片段预览生成方式必须为播放时自动生成或手动生成")
		}
		if typed.VideoMomentOpenMode != "dialog" && typed.VideoMomentOpenMode != "page" {
			return errors.New("片段收藏打开方式必须为弹窗播放器或详情播放页")
		}
		if typed.VideoPreviewInterval != 10 && typed.VideoPreviewInterval != 30 && typed.VideoPreviewInterval != 60 && typed.VideoPreviewInterval != 120 {
			return errors.New("默认抽图间隔必须为10、30、60或120秒")
		}
		if _, ok := value["casualViewModule"].(bool); !ok {
			return errors.New("随便看看开关必须为布尔值")
		}
		if n, ok := value["casualViewNumber"].(float64); !ok || n < 0 || n != float64(int(n)) {
			return errors.New("随便看看数量必须为非负整数")
		}
	case "import":
		var typed datatype.Config_ScanDisk
		if err := json.Unmarshal(data, &typed); err != nil {
			return fmt.Errorf("导入参数类型错误: %w", err)
		}
		if typed.ImportMode != "append" && typed.ImportMode != "cover" {
			return errors.New("无效的导入方式")
		}
		if _, ok := value["nfo"].(map[string]interface{}); !ok {
			return errors.New("NFO 配置必须为对象")
		}
		// 将嵌套 NFO 配置按定义重新编码，过滤未来或外部输入的未知成员。
		nfoData, _ := json.Marshal(typed.Nfo)
		var nfo map[string]interface{}
		_ = json.Unmarshal(nfoData, &nfo)
		value["nfo"] = nfo
		data, _ = json.Marshal(value)
	case "scraper":
		var typed datatype.Config_Scraper
		if err := json.Unmarshal(data, &typed); err != nil {
			return fmt.Errorf("刮削参数类型错误: %w", err)
		}
		if typed.Concurrency < 1 || typed.Concurrency > 10 || typed.Timeout < 1 || typed.Timeout > 300 || typed.RetryCount < 0 || typed.RetryCount > 10 {
			return errors.New("刮削并发、超时或重试次数超出允许范围")
		}
	}

	if revision == 0 {
		err := db.Create(&SharedLibraryConfig{Module: module, Revision: 1, Config: string(data)}).Error
		if err != nil {
			return fmt.Errorf("公共配置创建失败，请刷新后重试: %w", err)
		}
		return nil
	}
	result := db.Model(&SharedLibraryConfig{}).Where("module = ? AND revision = ?", module, revision).Updates(map[string]interface{}{"config": string(data), "revision": revision + 1})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("公共配置已被其他页面修改，请重新加载后再保存")
	}
	return nil
}

// SetLibraryConfigFollow 必须在事务中调用。开启时保存通用字段快照；退出可恢复或保留当前值。
func SetLibraryConfigFollow(db *gorm.DB, id, module string, following bool, revision int, detachModes ...string) error {
	state, err := SharedConfigStatus(db, id, module)
	if err != nil {
		return err
	}
	if !state.Available || state.Revision != revision {
		return errors.New("公共配置已变化或尚未创建，请刷新后重试")
	}
	var library FilesBases
	if err := db.Where("id = ?", id).Take(&library).Error; err != nil {
		return err
	}
	mode := "keep"
	if len(detachModes) > 0 && detachModes[0] != "" {
		mode = detachModes[0]
	}
	if mode != "keep" && mode != "restore" {
		return errors.New("不支持的退出跟随方式")
	}
	if following && state.Following {
		return nil
	}
	if !following && !state.Following {
		return nil
	}
	var setting FilesBasesSetting
	if err := db.Where("filesBases_id = ?", id).Take(&setting).Error; err != nil {
		return err
	}
	raw := setting.ConfigJsonData
	if module == "import" {
		raw = setting.ScanDiskJsonData
	}
	if module == "scraper" {
		raw = setting.ScraperJsonData
	}
	local, err := parseConfig(raw)
	if err != nil {
		return err
	}
	if following {
		data, err := json.Marshal(sharedFieldsOnly(module, local))
		if err != nil {
			return err
		}
		snapshot := string(data)
		return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&LibraryConfigFollow{FilesBasesID: id, Module: module, LocalSnapshot: &snapshot}).Error
	}
	var effective string
	if mode == "restore" {
		var follow LibraryConfigFollow
		if err := db.Where("filesBases_id = ? AND module = ?", id, module).Take(&follow).Error; err != nil {
			return err
		}
		if follow.LocalSnapshot == nil {
			return errors.New("没有跟随前的配置快照，请选择保留当前公共配置")
		}
		snapshot, err := parseConfig(*follow.LocalSnapshot)
		if err != nil {
			return err
		}
		// 只恢复通用参数；跟随期间修改的目录、演员选择等独立项保持当前值。
		for _, field := range SharedConfigFields(module) {
			delete(local, field)
			if value, ok := snapshot[field]; ok {
				local[field] = value
			}
		}
		data, err := json.Marshal(local)
		if err != nil {
			return err
		}
		effective = string(data)
	} else {
		effective, err = EffectiveLibraryConfig(db, id, module, raw)
		if err != nil {
			return err
		}
	}
	if err := db.Model(&FilesBasesSetting{}).Where("filesBases_id = ?", id).Update(sharedConfigColumn(module), effective).Error; err != nil {
		return err
	}
	return db.Where("filesBases_id = ? AND module = ?", id, module).Delete(&LibraryConfigFollow{}).Error
}
