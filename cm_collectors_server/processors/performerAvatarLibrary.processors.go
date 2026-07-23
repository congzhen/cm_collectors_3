package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/longbridgeapp/opencc"
	"gorm.io/gorm"
)

const (
	avatarLibraryDefaultBaseURL  = "https://raw.githubusercontent.com/gfriends/gfriends/master/"
	avatarLibraryFallbackBaseURL = "https://cdn.jsdelivr.net/gh/xinxin8816/gfriends/"
	avatarLibraryMaxFileSize     = 32 << 20
)

var avatarNumberSuffix = regexp.MustCompile(`-\d+$`)

var avatarNameConversion = struct {
	once                    sync.Once
	simplifiedToTraditional *opencc.OpenCC
	traditionalToSimplified *opencc.OpenCC
}{}

type PerformerAvatarLibrary struct{}

type performerAvatarInformation struct {
	TotalNum  any `json:"TotalNum"`
	TotalSize any `json:"TotalSize"`
	Timestamp any `json:"Timestamp"`
}

type orderedAvatarEntry struct {
	Name  string
	Value string
}

type orderedAvatarSource struct {
	Name    string
	Entries []orderedAvatarEntry
}

type orderedAvatarContent []orderedAvatarSource

// UnmarshalJSON 保留 Gfriends 文件树中的来源和图片顺序；推荐质量依赖该顺序，不能解码到普通 map。
func (content *orderedAvatarContent) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return errors.New("Gfriends Content 不是有效对象")
	}
	for decoder.More() {
		sourceNameToken, err := decoder.Token()
		if err != nil {
			return err
		}
		sourceName, ok := sourceNameToken.(string)
		if !ok {
			return errors.New("Gfriends 来源名称无效")
		}
		token, err = decoder.Token()
		if err != nil || token != json.Delim('{') {
			return fmt.Errorf("Gfriends 来源 %s 不是有效对象", sourceName)
		}
		source := orderedAvatarSource{Name: sourceName}
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			var value string
			if err := decoder.Decode(&value); err != nil {
				return err
			}
			source.Entries = append(source.Entries, orderedAvatarEntry{Name: nameToken.(string), Value: value})
		}
		if _, err := decoder.Token(); err != nil {
			return err
		}
		*content = append(*content, source)
	}
	_, err = decoder.Token()
	return err
}

type performerAvatarFileTree struct {
	Information performerAvatarInformation `json:"Information"`
	Content     orderedAvatarContent       `json:"Content"`
}

type PerformerAvatarCandidate struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	FileName string `json:"fileName"`
	AIFixed  bool   `json:"aiFixed"`
	Rank     int    `json:"rank"`
}

type PerformerAvatarLibraryStatus struct {
	Ready         bool                                   `json:"ready"`
	Updating      bool                                   `json:"updating"`
	FileSize      int64                                  `json:"fileSize"`
	UpdatedAt     string                                 `json:"updatedAt"`
	DataTimestamp string                                 `json:"dataTimestamp"`
	TotalNum      string                                 `json:"totalNum"`
	TotalSize     string                                 `json:"totalSize"`
	ActiveBaseURL string                                 `json:"activeBaseUrl"`
	Setting       datatype.PerformerAvatarLibrarySetting `json:"setting"`
}

type PerformerAvatarBatchPreview struct {
	Total              int `json:"total"`
	Matched            int `json:"matched"`
	Unmatched          int `json:"unmatched"`
	SkippedExisting    int `json:"skippedExisting"`
	MultipleCandidates int `json:"multipleCandidates"`
}

type PerformerAvatarBatchProgress struct {
	BatchID            string                        `json:"batchId"`
	Total              int                           `json:"total"`
	Completed          int                           `json:"completed"`
	Matched            int                           `json:"matched"`
	Success            int                           `json:"success"`
	Failed             int                           `json:"failed"`
	Unmatched          int                           `json:"unmatched"`
	SkippedExisting    int                           `json:"skippedExisting"`
	MultipleCandidates int                           `json:"multipleCandidates"`
	CurrentActors      []string                      `json:"currentActors"`
	Failures           []PerformerAvatarBatchFailure `json:"failures"`
	Done               bool                          `json:"done"`
}

type PerformerAvatarBatchFailure struct {
	PerformerID string `json:"performerId"`
	Name        string `json:"name"`
	Error       string `json:"error"`
}

type performerAvatarBatchProgressTracker struct {
	sync.RWMutex
	progress PerformerAvatarBatchProgress
	current  map[string]string
}

type PerformerAvatarBatchActor struct {
	ID               string `json:"id"`
	PerformerBasesID string `json:"performerBasesId"`
	Name             string `json:"name"`
	AliasName        string `json:"aliasName"`
	Photo            string `json:"photo"`
	HasPhoto         bool   `json:"hasPhoto"`
}

type PerformerAvatarBatchActorPage struct {
	DataList []PerformerAvatarBatchActor `json:"dataList"`
	Total    int64                       `json:"total"`
}

type performerAvatarMetadata struct {
	ActiveBaseURL string    `json:"activeBaseUrl"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type performerAvatarIndex struct {
	Tree     *performerAvatarFileTree
	ByName   map[string][]PerformerAvatarCandidate
	ByID     map[string]PerformerAvatarCandidate
	Metadata performerAvatarMetadata
}

var avatarLibraryRuntime = struct {
	sync.RWMutex
	Index    *performerAvatarIndex
	Updating bool
}{}

var avatarCandidateLocks sync.Map

var avatarBatchProgress sync.Map

func avatarLibraryDir() string {
	return filepath.Clean(core.GetConfig().PerformerAvatarLibrary.CachePath)
}

func avatarLibraryTreePath() string     { return filepath.Join(avatarLibraryDir(), "Filetree.json") }
func avatarLibraryMetadataPath() string { return filepath.Join(avatarLibraryDir(), "metadata.json") }

func (PerformerAvatarLibrary) Setting() datatype.PerformerAvatarLibrarySetting {
	appConfig := core.GetConfig()
	setting := datatype.PerformerAvatarLibrarySetting{
		CustomBaseURL:   appConfig.PerformerAvatarLibrary.CustomBaseURL,
		DefaultStrategy: datatype.PerformerAvatarStrategy(appConfig.PerformerAvatarLibrary.DefaultStrategy),
	}
	if !validAvatarStrategy(setting.DefaultStrategy) {
		setting.DefaultStrategy = datatype.PerformerAvatarStrategyRecommended
	}
	setting.CustomBaseURL = normalizeBaseURL(setting.CustomBaseURL)
	return setting
}

func normalizeAvatarLibrarySetting(setting datatype.PerformerAvatarLibrarySetting) (datatype.PerformerAvatarLibrarySetting, error) {
	if !validAvatarStrategy(setting.DefaultStrategy) {
		return setting, errors.New("无效的头像选择策略")
	}
	setting.CustomBaseURL = normalizeBaseURL(setting.CustomBaseURL)
	if setting.CustomBaseURL != "" {
		parsed, err := url.Parse(setting.CustomBaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return setting, errors.New("自定义镜像地址必须是有效的 HTTP 或 HTTPS 地址")
		}
	}
	return setting, nil
}

func (t PerformerAvatarLibrary) Status() PerformerAvatarLibraryStatus {
	setting := t.Setting()
	avatarLibraryRuntime.RLock()
	updating := avatarLibraryRuntime.Updating
	avatarLibraryRuntime.RUnlock()
	status := PerformerAvatarLibraryStatus{Setting: setting, Updating: updating}
	index, err := t.loadIndex()
	if err != nil {
		return status
	}
	status.Ready = true
	status.TotalNum = fmt.Sprint(index.Tree.Information.TotalNum)
	status.TotalSize = fmt.Sprint(index.Tree.Information.TotalSize)
	status.DataTimestamp = fmt.Sprint(index.Tree.Information.Timestamp)
	status.ActiveBaseURL = index.Metadata.ActiveBaseURL
	if info, err := os.Stat(avatarLibraryTreePath()); err == nil {
		status.FileSize = info.Size()
		status.UpdatedAt = info.ModTime().Format(time.RFC3339)
	}
	if !index.Metadata.UpdatedAt.IsZero() {
		status.UpdatedAt = index.Metadata.UpdatedAt.Format(time.RFC3339)
	}
	return status
}

func (t PerformerAvatarLibrary) UpdateDataFile() (PerformerAvatarLibraryStatus, error) {
	avatarLibraryRuntime.Lock()
	if avatarLibraryRuntime.Updating {
		avatarLibraryRuntime.Unlock()
		return t.Status(), errors.New("头像库数据正在更新")
	}
	avatarLibraryRuntime.Updating = true
	avatarLibraryRuntime.Unlock()
	defer func() {
		setAvatarLibraryUpdating(false)
	}()

	setting := t.Setting()
	bases := []string{}
	if setting.CustomBaseURL != "" {
		bases = append(bases, setting.CustomBaseURL)
	}
	bases = append(bases, avatarLibraryDefaultBaseURL, avatarLibraryFallbackBaseURL)
	client := &http.Client{Timeout: 90 * time.Second}
	var lastErr error
	for _, baseURL := range uniqueStrings(bases) {
		data, err := downloadAvatarLibraryFile(client, baseURL+"Filetree.json", avatarLibraryMaxFileSize)
		if err != nil {
			lastErr = err
			continue
		}
		index, err := buildAvatarLibraryIndex(data, performerAvatarMetadata{ActiveBaseURL: baseURL, UpdatedAt: core.TimeNow()})
		if err != nil {
			lastErr = err
			continue
		}
		if err := saveAvatarLibraryFiles(data, index.Metadata); err != nil {
			return t.Status(), err
		}
		avatarLibraryRuntime.Lock()
		avatarLibraryRuntime.Index = index
		avatarLibraryRuntime.Unlock()
		setAvatarLibraryUpdating(false)
		return t.Status(), nil
	}
	if lastErr == nil {
		lastErr = errors.New("没有可用的头像库数据源")
	}
	return t.Status(), lastErr
}

func (t PerformerAvatarLibrary) Candidates(performerID string) ([]PerformerAvatarCandidate, error) {
	performer, err := Performer{}.InfoByID(performerID)
	if err != nil {
		return nil, err
	}
	index, err := t.loadIndex()
	if err != nil {
		return nil, err
	}
	return matchAvatarCandidates(index, *performer), nil
}

func (t PerformerAvatarLibrary) CandidateImage(performerID, candidateID string) ([]byte, string, error) {
	candidates, err := t.Candidates(performerID)
	if err != nil {
		return nil, "", err
	}
	for _, candidate := range candidates {
		if candidate.ID == candidateID {
			return t.downloadCandidate(candidate)
		}
	}
	return nil, "", errors.New("头像候选不存在或不属于该演员")
}

func (t PerformerAvatarLibrary) Apply(performerID, candidateID string, overwrite bool) error {
	performer, err := Performer{}.InfoByID(performerID)
	if err != nil {
		return err
	}
	if performer.Photo != "" && !overwrite {
		return errors.New("演员已有头像，未允许覆盖")
	}
	index, err := t.loadIndex()
	if err != nil {
		return err
	}
	for _, candidate := range matchAvatarCandidates(index, *performer) {
		if candidate.ID == candidateID {
			return t.applyCandidate(*performer, candidate, overwrite)
		}
	}
	return errors.New("头像候选不存在或不属于该演员")
}

func (t PerformerAvatarLibrary) applyCandidate(performer models.Performer, candidate PerformerAvatarCandidate, overwrite bool) error {
	if performer.Photo != "" && !overwrite {
		return errors.New("演员已有头像，未允许覆盖")
	}
	data, _, err := t.downloadCandidate(candidate)
	if err != nil {
		return err
	}
	base64Data, err := utils.ImageBytesToBase64(data)
	if err != nil {
		return err
	}
	return (Performer{}).replacePhoto(performer, base64Data, overwrite)
}

func (t PerformerAvatarLibrary) BatchPreview(par datatype.ReqParam_PerformerAvatarBatch) (PerformerAvatarBatchPreview, error) {
	performers, err := t.batchPerformers(par.PerformerBasesID, par.PerformerIDs, par.AllPerformers)
	if err != nil {
		return PerformerAvatarBatchPreview{}, err
	}
	index, err := t.loadIndex()
	if err != nil {
		return PerformerAvatarBatchPreview{}, err
	}
	return previewAvatarBatch(index, performers, par.Overwrite), nil
}

func (t PerformerAvatarLibrary) StartBatchApply(par datatype.ReqParam_PerformerAvatarBatch) (PerformerAvatarBatchProgress, error) {
	if !validAvatarStrategy(par.Strategy) {
		return PerformerAvatarBatchProgress{}, errors.New("无效的头像选择策略")
	}
	performers, err := t.batchPerformers(par.PerformerBasesID, par.PerformerIDs, par.AllPerformers)
	if err != nil {
		return PerformerAvatarBatchProgress{}, err
	}
	index, err := t.loadIndex()
	if err != nil {
		return PerformerAvatarBatchProgress{}, err
	}
	preview := previewAvatarBatch(index, performers, par.Overwrite)
	batchID := core.GenerateUniqueID()
	tracker := startAvatarBatchProgress(batchID, preview)
	go t.runBatchApply(batchID, par, performers, index, tracker)
	progress, _ := (PerformerAvatarLibrary{}).BatchProgress(batchID)
	return progress, nil
}

func (t PerformerAvatarLibrary) runBatchApply(
	batchID string,
	par datatype.ReqParam_PerformerAvatarBatch,
	performers []models.Performer,
	index *performerAvatarIndex,
	tracker *performerAvatarBatchProgressTracker,
) {
	defer finishAvatarBatchProgress(batchID, tracker)
	jobs := make(chan models.Performer)
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for performer := range jobs {
				if performer.Photo != "" && !par.Overwrite {
					continue
				}
				candidate, ok := selectAvatarCandidate(matchAvatarCandidates(index, performer), par.Strategy)
				if !ok {
					continue
				}
				setAvatarBatchActorRunning(tracker, performer)
				err := t.applyCandidate(performer, candidate, par.Overwrite)
				completeAvatarBatchActor(tracker, performer, err)
			}
		}()
	}
	for _, performer := range performers {
		jobs <- performer
	}
	close(jobs)
	wg.Wait()
}

func (PerformerAvatarLibrary) BatchProgress(batchID string) (PerformerAvatarBatchProgress, bool) {
	value, ok := avatarBatchProgress.Load(batchID)
	if !ok {
		return PerformerAvatarBatchProgress{}, false
	}
	tracker := value.(*performerAvatarBatchProgressTracker)
	tracker.RLock()
	defer tracker.RUnlock()
	result := tracker.progress
	result.Failures = append(make([]PerformerAvatarBatchFailure, 0, len(tracker.progress.Failures)), tracker.progress.Failures...)
	result.CurrentActors = make([]string, 0, len(tracker.current))
	for _, name := range tracker.current {
		result.CurrentActors = append(result.CurrentActors, name)
	}
	sort.Strings(result.CurrentActors)
	return result, true
}

func startAvatarBatchProgress(batchID string, preview PerformerAvatarBatchPreview) *performerAvatarBatchProgressTracker {
	if strings.TrimSpace(batchID) == "" {
		return nil
	}
	tracker := &performerAvatarBatchProgressTracker{
		progress: PerformerAvatarBatchProgress{
			BatchID: batchID, Total: preview.Total,
			Completed: preview.Unmatched + preview.SkippedExisting,
			Matched:   preview.Matched, Unmatched: preview.Unmatched,
			SkippedExisting: preview.SkippedExisting, MultipleCandidates: preview.MultipleCandidates,
		},
		current: map[string]string{},
	}
	avatarBatchProgress.Store(batchID, tracker)
	return tracker
}

func setAvatarBatchActorRunning(tracker *performerAvatarBatchProgressTracker, performer models.Performer) {
	if tracker == nil {
		return
	}
	tracker.Lock()
	tracker.current[performer.ID] = performer.Name
	tracker.Unlock()
}

func completeAvatarBatchActor(tracker *performerAvatarBatchProgressTracker, performer models.Performer, applyErr error) {
	if tracker == nil {
		return
	}
	tracker.Lock()
	delete(tracker.current, performer.ID)
	tracker.progress.Completed++
	if applyErr == nil {
		tracker.progress.Success++
	} else {
		tracker.progress.Failed++
		tracker.progress.Failures = append(tracker.progress.Failures, PerformerAvatarBatchFailure{
			PerformerID: performer.ID, Name: performer.Name, Error: applyErr.Error(),
		})
	}
	tracker.Unlock()
}

func finishAvatarBatchProgress(batchID string, tracker *performerAvatarBatchProgressTracker) {
	if tracker == nil {
		return
	}
	tracker.Lock()
	tracker.current = map[string]string{}
	tracker.progress.Completed = tracker.progress.Total
	tracker.progress.Done = true
	tracker.Unlock()
	time.AfterFunc(10*time.Minute, func() {
		avatarBatchProgress.CompareAndDelete(batchID, tracker)
	})
}

func (PerformerAvatarLibrary) BatchActors(performerBasesID string, page, limit int, search, photoFilter string) (PerformerAvatarBatchActorPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 90
	}
	query := avatarBatchActorQuery(performerBasesID, search, photoFilter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return PerformerAvatarBatchActorPage{}, err
	}
	var performers []models.Performer
	err := query.Select("id", "performerBases_id", "name", "aliasName", "photo").
		Order("addTime desc").Offset((page - 1) * limit).Limit(limit).Find(&performers).Error
	if err != nil {
		return PerformerAvatarBatchActorPage{}, err
	}
	result := make([]PerformerAvatarBatchActor, 0, len(performers))
	for _, performer := range performers {
		result = append(result, PerformerAvatarBatchActor{
			ID: performer.ID, PerformerBasesID: performer.PerformerBasesID, Name: performer.Name,
			AliasName: performer.AliasName, Photo: performer.Photo, HasPhoto: performer.Photo != "",
		})
	}
	return PerformerAvatarBatchActorPage{DataList: result, Total: total}, nil
}

func (PerformerAvatarLibrary) BatchActorIDs(performerBasesID, search, photoFilter string) ([]string, error) {
	var ids []string
	err := avatarBatchActorQuery(performerBasesID, search, photoFilter).Pluck("id", &ids).Error
	return ids, err
}

func avatarBatchActorQuery(performerBasesID, search, photoFilter string) *gorm.DB {
	query := core.DBS().Model(&models.Performer{}).
		Where("performerBases_id = ? AND status = ?", performerBasesID, true)
	if search = strings.TrimSpace(search); search != "" {
		keyword := "%" + search + "%"
		query = query.Where("(name LIKE ? OR aliasName LIKE ?)", keyword, keyword)
	}
	switch photoFilter {
	case "missing":
		query = query.Where("(photo = '' OR photo IS NULL)")
	case "existing":
		query = query.Where("photo <> ''")
	}
	return query
}

func (PerformerAvatarLibrary) batchPerformers(performerBasesID string, performerIDs []string, allPerformers bool) ([]models.Performer, error) {
	var performers []models.Performer
	if !allPerformers && len(performerIDs) == 0 {
		return performers, nil
	}
	query := core.DBS().Where("performerBases_id = ? AND status = ?", performerBasesID, true)
	if !allPerformers {
		query = query.Where("id IN ?", performerIDs)
	}
	err := query.Order("addTime desc").Find(&performers).Error
	return performers, err
}

func (t PerformerAvatarLibrary) loadIndex() (*performerAvatarIndex, error) {
	avatarLibraryRuntime.RLock()
	index := avatarLibraryRuntime.Index
	avatarLibraryRuntime.RUnlock()
	if index != nil {
		return index, nil
	}
	data, err := os.ReadFile(avatarLibraryTreePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("尚未下载演员头像库数据文件")
		}
		return nil, err
	}
	metadata := performerAvatarMetadata{ActiveBaseURL: avatarLibraryDefaultBaseURL}
	if metadataData, err := os.ReadFile(avatarLibraryMetadataPath()); err == nil {
		_ = json.Unmarshal(metadataData, &metadata)
	}
	index, err = buildAvatarLibraryIndex(data, metadata)
	if err != nil {
		return nil, err
	}
	avatarLibraryRuntime.Lock()
	avatarLibraryRuntime.Index = index
	avatarLibraryRuntime.Unlock()
	return index, nil
}

func buildAvatarLibraryIndex(data []byte, metadata performerAvatarMetadata) (*performerAvatarIndex, error) {
	var tree performerAvatarFileTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("头像库数据文件解析失败: %w", err)
	}
	if len(tree.Content) == 0 {
		return nil, errors.New("头像库数据文件中没有 Content 数据")
	}
	index := &performerAvatarIndex{Tree: &tree, ByName: map[string][]PerformerAvatarCandidate{}, ByID: map[string]PerformerAvatarCandidate{}, Metadata: metadata}
	rank := 0
	for _, source := range tree.Content {
		for _, entry := range source.Entries {
			rank++
			lookupName := strings.TrimSuffix(entry.Name, filepath.Ext(entry.Name))
			lookupName = avatarNumberSuffix.ReplaceAllString(lookupName, "")
			lookupName = normalizeAvatarName(lookupName)
			if lookupName == "" || entry.Value == "" {
				continue
			}
			candidate := PerformerAvatarCandidate{ID: avatarCandidateID(source.Name, entry.Value), Source: source.Name, FileName: entry.Value, AIFixed: strings.Contains(entry.Value, "AI-Fix-"), Rank: rank}
			index.ByName[lookupName] = appendUniqueCandidate(index.ByName[lookupName], candidate)
			index.ByID[candidate.ID] = candidate
		}
	}
	if len(index.ByName) == 0 {
		return nil, errors.New("头像库数据文件没有可检索的头像")
	}
	return index, nil
}

func matchAvatarCandidates(index *performerAvatarIndex, performer models.Performer) []PerformerAvatarCandidate {
	// 主姓名拥有最高优先级。仅当主姓名没有命中时，才把别名按常见
	// 中英文分隔符拆开逐一匹配，避免别名候选覆盖正确的主姓名结果。
	result := matchAvatarCandidateNames(index, []string{performer.Name})
	if len(result) > 0 {
		return result
	}
	aliases := splitAvatarAliases(performer.AliasName)
	result = matchAvatarCandidateNames(index, aliases)
	if len(result) > 0 {
		return result
	}

	// 精确匹配均失败后再尝试简繁体变体，仍保持主姓名优先。
	result = matchAvatarCandidateNames(index, convertedAvatarNameVariants(performer.Name))
	if len(result) > 0 {
		return result
	}
	convertedAliases := make([]string, 0, len(aliases)*2)
	for _, alias := range aliases {
		convertedAliases = append(convertedAliases, convertedAvatarNameVariants(alias)...)
	}
	return matchAvatarCandidateNames(index, convertedAliases)
}

func matchAvatarCandidateNames(index *performerAvatarIndex, names []string) []PerformerAvatarCandidate {
	result := []PerformerAvatarCandidate{}
	for _, name := range names {
		for _, candidate := range index.ByName[normalizeAvatarName(name)] {
			result = appendUniqueCandidate(result, candidate)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Rank > result[j].Rank })
	return result
}

func convertedAvatarNameVariants(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	avatarNameConversion.once.Do(func() {
		avatarNameConversion.simplifiedToTraditional, _ = opencc.New("s2t")
		avatarNameConversion.traditionalToSimplified, _ = opencc.New("t2s")
	})

	originalKey := normalizeAvatarName(name)
	result := []string{}
	seen := map[string]bool{originalKey: true}
	for _, converter := range []*opencc.OpenCC{
		avatarNameConversion.simplifiedToTraditional,
		avatarNameConversion.traditionalToSimplified,
	} {
		if converter == nil {
			continue
		}
		converted, err := converter.Convert(name)
		key := normalizeAvatarName(converted)
		if err == nil && key != "" && !seen[key] {
			seen[key] = true
			result = append(result, converted)
		}
	}
	return result
}

func selectAvatarCandidate(candidates []PerformerAvatarCandidate, strategy datatype.PerformerAvatarStrategy) (PerformerAvatarCandidate, bool) {
	if len(candidates) == 0 {
		return PerformerAvatarCandidate{}, false
	}
	wantAI := strategy == datatype.PerformerAvatarStrategyAIFix
	if strategy != datatype.PerformerAvatarStrategyRecommended {
		for _, candidate := range candidates {
			if candidate.AIFixed == wantAI {
				return candidate, true
			}
		}
	}
	return candidates[0], true
}

func previewAvatarBatch(index *performerAvatarIndex, performers []models.Performer, overwrite bool) PerformerAvatarBatchPreview {
	result := PerformerAvatarBatchPreview{Total: len(performers)}
	for _, performer := range performers {
		if performer.Photo != "" && !overwrite {
			result.SkippedExisting++
			continue
		}
		candidates := matchAvatarCandidates(index, performer)
		if len(candidates) == 0 {
			result.Unmatched++
			continue
		}
		result.Matched++
		if len(candidates) > 1 {
			result.MultipleCandidates++
		}
	}
	return result
}

func (t PerformerAvatarLibrary) downloadCandidate(candidate PerformerAvatarCandidate) ([]byte, string, error) {
	lockValue, _ := avatarCandidateLocks.LoadOrStore(candidate.ID, &sync.Mutex{})
	candidateLock := lockValue.(*sync.Mutex)
	candidateLock.Lock()
	defer candidateLock.Unlock()

	cachePath := filepath.Join(avatarLibraryDir(), "images", candidate.ID+".cache")
	if data, err := os.ReadFile(cachePath); err == nil {
		contentType := http.DetectContentType(data)
		if strings.HasPrefix(contentType, "image/") {
			return data, contentType, nil
		}
	}

	index, err := t.loadIndex()
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{Timeout: 45 * time.Second}
	setting := t.Setting()
	bases := uniqueStrings([]string{index.Metadata.ActiveBaseURL, setting.CustomBaseURL, avatarLibraryDefaultBaseURL, avatarLibraryFallbackBaseURL})
	var lastErr error
	for _, baseURL := range bases {
		fileURL, err := avatarCandidateURL(baseURL, candidate)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := downloadAvatarLibraryFileWithRetry(client, fileURL, 20<<20, 3)
		if err != nil {
			lastErr = err
			continue
		}
		contentType := http.DetectContentType(data)
		if !strings.HasPrefix(contentType, "image/") {
			lastErr = errors.New("下载内容不是有效图片")
			continue
		}
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err == nil {
			tempPath := cachePath + ".tmp"
			if err := os.WriteFile(tempPath, data, 0644); err == nil {
				_ = replaceFileSafely(tempPath, cachePath)
			}
		}
		return data, contentType, nil
	}
	return nil, "", lastErr
}

type avatarLibraryHTTPError struct {
	StatusCode int
}

func (err avatarLibraryHTTPError) Error() string {
	return fmt.Sprintf("头像库返回 HTTP %d", err.StatusCode)
}

func downloadAvatarLibraryFileWithRetry(client *http.Client, fileURL string, maxSize int64, attempts int) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		data, err := downloadAvatarLibraryFile(client, fileURL, maxSize)
		if err == nil {
			return data, nil
		}
		lastErr = err
		var httpErr avatarLibraryHTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode != http.StatusTooManyRequests && httpErr.StatusCode < 500 {
			break
		}
		if attempt+1 < attempts {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	return nil, lastErr
}

func downloadAvatarLibraryFile(client *http.Client, fileURL string, maxSize int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CM-Collectors-3/Gfriends-Avatar-Library")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, avatarLibraryHTTPError{StatusCode: response.StatusCode}
	}
	if response.ContentLength > maxSize {
		return nil, errors.New("下载文件超过允许大小")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, errors.New("下载文件超过允许大小")
	}
	return data, nil
}

func saveAvatarLibraryFiles(data []byte, metadata performerAvatarMetadata) error {
	if err := os.MkdirAll(avatarLibraryDir(), 0755); err != nil {
		return err
	}
	tempPath := avatarLibraryTreePath() + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}
	if err := replaceFileSafely(tempPath, avatarLibraryTreePath()); err != nil {
		return err
	}
	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	return os.WriteFile(avatarLibraryMetadataPath(), metadataData, 0644)
}

func replaceFileSafely(tempPath, targetPath string) error {
	backupPath := targetPath + ".bak"
	_ = os.Remove(backupPath)
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, backupPath); err != nil {
			return err
		}
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Rename(backupPath, targetPath)
		return err
	}
	_ = os.Remove(backupPath)
	return nil
}

func avatarCandidateURL(baseURL string, candidate PerformerAvatarCandidate) (string, error) {
	baseURL = normalizeBaseURL(baseURL)
	if baseURL == "" {
		baseURL = avatarLibraryDefaultBaseURL
	}
	fileName, query, _ := strings.Cut(candidate.FileName, "?")
	result := baseURL + "Content/" + url.PathEscape(candidate.Source) + "/" + url.PathEscape(fileName)
	if query != "" {
		result += "?" + query
	}
	return result, nil
}

func avatarCandidateID(source, fileName string) string {
	sum := sha256.Sum256([]byte(source + "\n" + fileName))
	return hex.EncodeToString(sum[:12])
}

func normalizeAvatarName(name string) string {
	name = strings.ReplaceAll(name, "　", " ")
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}

func splitAvatarAliases(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Keep the original alias as a candidate so names containing spaces (for
	// example an English full name) can still be matched. Add split candidates
	// to tolerate mixed punctuation and whitespace used by imported metadata.
	result := make([]string, 0, 8)
	seen := map[string]bool{}
	appendAlias := func(alias string) {
		alias = strings.TrimSpace(alias)
		key := normalizeAvatarName(alias)
		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, alias)
		}
	}
	appendAlias(value)

	parts := strings.FieldsFunc(value, func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		}
		return strings.ContainsRune(",，、;；:：/／\\|｜·・()（）[]【】{}<>《》“”\"'‘’", r)
	})
	for _, part := range parts {
		appendAlias(part)
	}
	return result
}

func appendUniqueCandidate(list []PerformerAvatarCandidate, candidate PerformerAvatarCandidate) []PerformerAvatarCandidate {
	for _, existing := range list {
		if existing.ID == candidate.ID {
			return list
		}
	}
	return append(list, candidate)
}

func validAvatarStrategy(strategy datatype.PerformerAvatarStrategy) bool {
	return strategy == datatype.PerformerAvatarStrategyRecommended || strategy == datatype.PerformerAvatarStrategyOriginal || strategy == datatype.PerformerAvatarStrategyAIFix
}

func normalizeBaseURL(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && !strings.HasSuffix(value, "/") {
		value += "/"
	}
	return value
}

func uniqueStrings(values []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = normalizeBaseURL(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func setAvatarLibraryUpdating(updating bool) {
	avatarLibraryRuntime.Lock()
	avatarLibraryRuntime.Updating = updating
	avatarLibraryRuntime.Unlock()
}
