package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"fmt"
	"math"
	"math/bits"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	"gorm.io/gorm"
)

type VideoFingerprint struct{}

var fingerprintPositions = []float64{0.05, 0.15, 0.25, 0.35, 0.45, 0.55, 0.65, 0.75, 0.85, 0.95}

const fingerprintConcurrency = 2
const fingerprintInsertBatchSize = 200

type VideoFingerprintTaskStatus struct {
	Running        bool   `json:"running"`
	FilesBasesID   string `json:"files_bases_id"`
	StartedAt      string `json:"started_at"`
	LastFinishedAt string `json:"last_finished_at"`
	LastError      string `json:"last_error"`
	LastSuccess    int    `json:"last_success"`
	LastFailed     int    `json:"last_failed"`
}

var fingerprintTaskMu sync.Mutex
var fingerprintTaskStatus = VideoFingerprintTaskStatus{}

func (VideoFingerprint) InsertPendingIfNotExists(dramaSeriesID, resourcesID, filesBasesID, src string) error {
	dbs := core.DBS()
	vfM := models.VideoFingerprint{}
	if vfM.ExistsByDramaSeriesID(dbs, dramaSeriesID) {
		return nil
	}
	return vfM.Create(dbs, &models.VideoFingerprint{
		ID:            core.GenerateUniqueID(),
		DramaSeriesID: dramaSeriesID,
		ResourcesID:   resourcesID,
		FilesBasesID:  filesBasesID,
		Src:           src,
		Status:        models.VideoFingerprintStatus_Pending,
	})
}

func (VideoFingerprint) InsertPendingBatch(dramaSeriesList []models.ResourcesDramaSeries, resourcesID, filesBasesID string) error {
	dbs := core.DBS()
	vfM := models.VideoFingerprint{}
	toInsert := make([]models.VideoFingerprint, 0, len(dramaSeriesList))
	for _, ds := range dramaSeriesList {
		if vfM.ExistsByDramaSeriesID(dbs, ds.ID) {
			continue
		}
		toInsert = append(toInsert, models.VideoFingerprint{
			ID:            core.GenerateUniqueID(),
			DramaSeriesID: ds.ID,
			ResourcesID:   resourcesID,
			FilesBasesID:  filesBasesID,
			Src:           ds.Src,
			Status:        models.VideoFingerprintStatus_Pending,
		})
	}
	return vfM.Creates(dbs, &toInsert)
}

func (VideoFingerprint) DeleteByResourcesID(resourcesID string) error {
	return models.VideoFingerprint{}.DeleteByResourcesID(core.DBS(), resourcesID)
}

func (VideoFingerprint) DeleteByDramaSeriesIDs(dramaSeriesIDs []string) error {
	return models.VideoFingerprint{}.DeleteByDramaSeriesIDs(core.DBS(), dramaSeriesIDs)
}

func (VideoFingerprint) Stats(filesBasesID string) (*models.VideoFingerprintStats, error) {
	return models.VideoFingerprint{}.Stats(core.DBS(), filesBasesID)
}

func (VideoFingerprint) ReScanMissingFingerprints() (int, error) {
	return VideoFingerprint{}.ReScanMissingFingerprintsByFilesBasesID("")
}

func (VideoFingerprint) ReScanMissingFingerprintsByFilesBasesID(filesBasesID string) (int, error) {
	dbs := core.DBS()
	vfM := models.VideoFingerprint{}
	missingList, err := vfM.MissingVideoDramaSeries(dbs, filesBasesID)
	if err != nil {
		return 0, err
	}

	toInsert := make([]models.VideoFingerprint, 0, len(missingList))
	for _, ds := range missingList {
		toInsert = append(toInsert, models.VideoFingerprint{
			ID:            core.GenerateUniqueID(),
			DramaSeriesID: ds.DramaSeriesID,
			ResourcesID:   ds.ResourcesID,
			FilesBasesID:  ds.FilesBasesID,
			Src:           ds.Src,
			Status:        models.VideoFingerprintStatus_Pending,
		})
	}
	for start := 0; start < len(toInsert); start += fingerprintInsertBatchSize {
		end := start + fingerprintInsertBatchSize
		if end > len(toInsert) {
			end = len(toInsert)
		}
		batch := toInsert[start:end]
		if err := vfM.Creates(dbs, &batch); err != nil {
			return 0, err
		}
	}
	return len(toInsert), nil
}

func (t VideoFingerprint) ResetAllFingerprints() error {
	dbs := core.DBS()
	vfM := models.VideoFingerprint{}
	if err := vfM.DeleteAll(dbs); err != nil {
		return err
	}
	_, err := t.ReScanMissingFingerprints()
	return err
}

func (t VideoFingerprint) ResetFailed(filesBasesID string) (int64, error) {
	return models.VideoFingerprint{}.ResetFailed(core.DBS(), filesBasesID)
}

func (t VideoFingerprint) TaskStatus() VideoFingerprintTaskStatus {
	fingerprintTaskMu.Lock()
	defer fingerprintTaskMu.Unlock()
	return fingerprintTaskStatus
}

func (t VideoFingerprint) ComputePendingFingerprints(batchSize int) error {
	return t.ComputePendingFingerprintsByFilesBasesID("", batchSize)
}

func (t VideoFingerprint) ComputePendingFingerprintsByFilesBasesID(filesBasesID string, batchSize int) error {
	if !startFingerprintTask(filesBasesID) {
		return fmt.Errorf("视频指纹任务正在执行中")
	}
	defer finishFingerprintTask()
	return t.computeAllPendingFingerprintsLocked(filesBasesID, batchSize)
}

func (t VideoFingerprint) StartComputePendingFingerprintsAsync(filesBasesID string, batchSize int) error {
	if !startFingerprintTask(filesBasesID) {
		return fmt.Errorf("视频指纹任务正在执行中")
	}
	go func() {
		defer finishFingerprintTask()
		_ = t.computeAllPendingFingerprintsLocked(filesBasesID, batchSize)
	}()
	return nil
}

func (t VideoFingerprint) ComputeWithReScanByFilesBasesID(filesBasesID string, batchSize int) error {
	if !startFingerprintTask(filesBasesID) {
		return fmt.Errorf("视频指纹任务正在执行中")
	}
	defer finishFingerprintTask()
	if _, err := t.ReScanMissingFingerprintsByFilesBasesID(filesBasesID); err != nil {
		setFingerprintTaskError(err)
		return err
	}
	return t.computeAllPendingFingerprintsLocked(filesBasesID, batchSize)
}

func startFingerprintTask(filesBasesID string) bool {
	fingerprintTaskMu.Lock()
	defer fingerprintTaskMu.Unlock()
	if fingerprintTaskStatus.Running {
		return false
	}
	fingerprintTaskStatus.Running = true
	fingerprintTaskStatus.FilesBasesID = filesBasesID
	fingerprintTaskStatus.StartedAt = time.Now().Format(time.DateTime)
	fingerprintTaskStatus.LastError = ""
	fingerprintTaskStatus.LastSuccess = 0
	fingerprintTaskStatus.LastFailed = 0
	return true
}

func finishFingerprintTask() {
	fingerprintTaskMu.Lock()
	fingerprintTaskStatus.Running = false
	fingerprintTaskStatus.LastFinishedAt = time.Now().Format(time.DateTime)
	fingerprintTaskMu.Unlock()
}

func (t VideoFingerprint) computeAllPendingFingerprintsLocked(filesBasesID string, batchSize int) error {
	for {
		hasMore, err := t.computePendingFingerprintsBatchLocked(filesBasesID, batchSize)
		if err != nil {
			return err
		}
		if !hasMore {
			return nil
		}
	}
}

func (t VideoFingerprint) computePendingFingerprintsBatchLocked(filesBasesID string, batchSize int) (bool, error) {
	dbs := core.DBS()
	pending, err := models.VideoFingerprint{}.ListPending(dbs, filesBasesID, batchSize)
	if err != nil {
		setFingerprintTaskError(err)
		return false, err
	}
	if len(*pending) == 0 {
		return false, nil
	}

	sem := make(chan struct{}, fingerprintConcurrency)
	var wg sync.WaitGroup
	var countMu sync.Mutex
	successCount := 0
	failedCount := 0
	for _, vf := range *pending {
		wg.Add(1)
		sem <- struct{}{}
		go func(record models.VideoFingerprint) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := t.computeOne(&record); err != nil {
				fmt.Printf("视频指纹计算失败 [%s]: %v\n", record.DramaSeriesID, err)
				setFingerprintTaskError(err)
				countMu.Lock()
				failedCount++
				countMu.Unlock()
				return
			}
			countMu.Lock()
			if record.Status == models.VideoFingerprintStatus_Done {
				successCount++
			} else {
				failedCount++
			}
			countMu.Unlock()
		}(vf)
	}
	wg.Wait()
	setFingerprintTaskCount(successCount, failedCount)
	return len(*pending) >= batchSize, nil
}

func setFingerprintTaskError(err error) {
	fingerprintTaskMu.Lock()
	defer fingerprintTaskMu.Unlock()
	fingerprintTaskStatus.LastError = err.Error()
}

func setFingerprintTaskCount(successCount, failedCount int) {
	fingerprintTaskMu.Lock()
	defer fingerprintTaskMu.Unlock()
	fingerprintTaskStatus.LastSuccess += successCount
	fingerprintTaskStatus.LastFailed += failedCount
}

func (VideoFingerprint) computeOne(vf *models.VideoFingerprint) error {
	dbs := core.DBS()

	if _, err := os.Stat(vf.Src); os.IsNotExist(err) {
		vf.Status = models.VideoFingerprintStatus_Failed
		vf.FailReason = "文件不存在"
		return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
	}

	formatInfo, err := processorsffmpeg.VideoInfo{}.GetVideoFormatInfo(vf.Src)
	if err != nil {
		vf.Status = models.VideoFingerprintStatus_Failed
		vf.FailReason = fmt.Sprintf("获取视频信息失败: %v", err)
		return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
	}

	durationStr := formatInfo.Format.Duration
	if durationStr == "" {
		for _, s := range formatInfo.Streams {
			if s.CodecType == "video" && s.Duration != "" {
				durationStr = s.Duration
				break
			}
		}
	}
	duration, _ := strconv.ParseFloat(durationStr, 64)
	if duration < 1 {
		vf.Status = models.VideoFingerprintStatus_Failed
		vf.FailReason = "视频时长过短"
		return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
	}

	images, err := processorsffmpeg.FrameExtract{}.ExtractFramesAtPositions(vf.Src, duration, fingerprintPositions)
	if err != nil {
		vf.Status = models.VideoFingerprintStatus_Failed
		vf.FailReason = fmt.Sprintf("提取帧失败: %v", err)
		return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
	}

	hashes := make([]string, len(images))
	for i, img := range images {
		h, err := goimagehash.PerceptionHash(img)
		if err != nil {
			vf.Status = models.VideoFingerprintStatus_Failed
			vf.FailReason = fmt.Sprintf("计算pHash失败: %v", err)
			return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
		}
		hashes[i] = fmt.Sprintf("%016x", h.GetHash())
	}

	vf.Duration = duration
	vf.PHash05 = hashes[0]
	vf.PHash15 = hashes[1]
	vf.PHash25 = hashes[2]
	vf.PHash35 = hashes[3]
	vf.PHash45 = hashes[4]
	vf.PHash55 = hashes[5]
	vf.PHash65 = hashes[6]
	vf.PHash75 = hashes[7]
	vf.PHash85 = hashes[8]
	vf.PHash95 = hashes[9]
	vf.Status = models.VideoFingerprintStatus_Done
	vf.FailReason = ""

	return models.VideoFingerprint{}.UpdateComputed(dbs, vf)
}

func hammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

func parseHash(hash string) (uint64, bool) {
	value, err := strconv.ParseUint(hash, 16, 64)
	return value, err == nil
}

func matchModeHashes(vf models.VideoFingerprint, mode string) []string {
	switch mode {
	case "minimal":
		return []string{vf.PHash05}
	case "loose":
		return []string{vf.PHash25, vf.PHash35, vf.PHash55, vf.PHash65, vf.PHash75}
	case "high":
		return []string{vf.PHash15, vf.PHash25, vf.PHash35, vf.PHash45, vf.PHash55, vf.PHash65, vf.PHash75}
	default:
		return []string{vf.PHash05, vf.PHash15, vf.PHash25, vf.PHash35, vf.PHash45,
			vf.PHash55, vf.PHash65, vf.PHash75, vf.PHash85, vf.PHash95}
	}
}

func hashDistances(a, b []string) ([]int, bool) {
	if len(a) == 0 || len(a) != len(b) {
		return nil, false
	}
	distances := make([]int, 0, len(a))
	for i := range a {
		av, ok := parseHash(a[i])
		if !ok {
			return nil, false
		}
		bv, ok := parseHash(b[i])
		if !ok {
			return nil, false
		}
		distances = append(distances, hammingDistance(av, bv))
	}
	return distances, true
}

func matchFingerprints(a, b models.VideoFingerprint, mode string, threshold int, durationTolerance float64) (float64, bool) {
	if durationTolerance > 0 && a.Duration > 0 && b.Duration > 0 && math.Abs(a.Duration-b.Duration) > durationTolerance {
		return 64, false
	}
	distances, ok := hashDistances(matchModeHashes(a, mode), matchModeHashes(b, mode))
	if !ok {
		return 64, false
	}
	total := 0
	for _, d := range distances {
		if d > threshold {
			return 64, false
		}
		total += d
	}
	return float64(total) / float64(len(distances)), true
}

func durationBucket(duration, durationTolerance float64) int {
	if durationTolerance <= 0 || duration <= 0 {
		return 0
	}
	return int(math.Floor(duration / durationTolerance))
}

func buildDurationBuckets(items []models.VideoFingerprint, durationTolerance float64) map[int][]int {
	buckets := map[int][]int{}
	if durationTolerance <= 0 {
		for i := range items {
			buckets[0] = append(buckets[0], i)
		}
		return buckets
	}
	for i, item := range items {
		bucket := durationBucket(item.Duration, durationTolerance)
		buckets[bucket] = append(buckets[bucket], i)
	}
	return buckets
}

func candidateIndices(currentIndex int, items []models.VideoFingerprint, buckets map[int][]int, durationTolerance float64) []int {
	if durationTolerance <= 0 {
		candidates := make([]int, 0, len(items)-currentIndex-1)
		for i := currentIndex + 1; i < len(items); i++ {
			candidates = append(candidates, i)
		}
		return candidates
	}
	bucket := durationBucket(items[currentIndex].Duration, durationTolerance)
	candidates := []int{}
	seen := map[int]bool{}
	for _, b := range []int{bucket - 1, bucket, bucket + 1} {
		for _, idx := range buckets[b] {
			if idx <= currentIndex || seen[idx] {
				continue
			}
			seen[idx] = true
			candidates = append(candidates, idx)
		}
	}
	return candidates
}

type DuplicateGroup struct {
	Items        []DuplicateItem `json:"items"`
	AvgDistance  float64         `json:"avg_distance"`
	MatchedCount int             `json:"matched_count"`
}

type DuplicateItem struct {
	DramaSeriesID string  `json:"drama_series_id"`
	ResourcesID   string  `json:"resources_id"`
	FilesBasesID  string  `json:"files_bases_id"`
	Src           string  `json:"src"`
	ResourceTitle string  `json:"resource_title"`
	CoverPoster   string  `json:"cover_poster"`
	Duration      float64 `json:"duration"`
}

func (t VideoFingerprint) QueryDuplicates(filesBasesID, matchMode string, threshold int, durationTolerance float64, page, limit int) ([]DuplicateGroup, int, error) {
	dbs := core.DBS()
	list, err := models.VideoFingerprint{}.ListDone(dbs, filesBasesID)
	if err != nil {
		return nil, 0, err
	}
	if len(*list) == 0 {
		return []DuplicateGroup{}, 0, nil
	}

	resMap := map[string]models.Resources{}
	var resList []models.Resources
	dbs.Select("id, title, coverPoster, filesBases_id").Find(&resList)
	for _, r := range resList {
		resMap[r.ID] = r
	}

	items := *list
	used := make([]bool, len(items))
	allGroups := []DuplicateGroup{}
	buckets := buildDurationBuckets(items, durationTolerance)

	for i := range items {
		if used[i] {
			continue
		}
		indices := []int{i}
		distances := []float64{}
		for _, j := range candidateIndices(i, items, buckets, durationTolerance) {
			if used[j] {
				continue
			}
			matchesAll := true
			candidateDistances := []float64{}
			for _, idx := range indices {
				d, matched := matchFingerprints(items[idx], items[j], matchMode, threshold, durationTolerance)
				if !matched {
					matchesAll = false
					break
				}
				candidateDistances = append(candidateDistances, d)
			}
			if matchesAll {
				indices = append(indices, j)
				distances = append(distances, candidateDistances...)
				used[j] = true
			}
		}
		if len(indices) < 2 {
			continue
		}
		used[i] = true
		group := DuplicateGroup{
			Items:        make([]DuplicateItem, 0, len(indices)),
			MatchedCount: len(matchModeHashes(items[i], matchMode)),
		}
		if len(distances) > 0 {
			sum := 0.0
			for _, d := range distances {
				sum += d
			}
			group.AvgDistance = sum / float64(len(distances))
		}
		for _, idx := range indices {
			vf := items[idx]
			res := resMap[vf.ResourcesID]
			group.Items = append(group.Items, DuplicateItem{
				DramaSeriesID: vf.DramaSeriesID,
				ResourcesID:   vf.ResourcesID,
				FilesBasesID:  vf.FilesBasesID,
				Src:           vf.Src,
				ResourceTitle: res.Title,
				CoverPoster:   res.CoverPoster,
				Duration:      vf.Duration,
			})
		}
		allGroups = append(allGroups, group)
	}

	total := len(allGroups)
	start := (page - 1) * limit
	if start >= total {
		return []DuplicateGroup{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return allGroups[start:end], total, nil
}

func (VideoFingerprint) DeleteDramaSeries(dramaSeriesIDs []string, deleteFile bool) error {
	db := core.DBS()
	pathsToDelete := []string{}
	for _, dramaSeriesID := range dramaSeriesIDs {
		info, err := models.ResourcesDramaSeries{}.Info(db, dramaSeriesID)
		if err != nil {
			return err
		}
		list, err := models.ResourcesDramaSeries{}.ListByResourceID(db, info.ResourcesID)
		if err != nil {
			return err
		}
		if deleteFile && info.Src != "" {
			pathsToDelete = append(pathsToDelete, info.Src)
		}
		if len(*list) <= 1 {
			resP := Resources{}
			if err := resP.DeleteResource(info.ResourcesID); err != nil {
				return err
			}
			continue
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			vfM := models.VideoFingerprint{}
			if err := vfM.DeleteByDramaSeriesID(tx, dramaSeriesID); err != nil {
				return err
			}
			dsM := models.ResourcesDramaSeries{}
			return dsM.DeleteIDS(tx, []string{dramaSeriesID})
		})
		if err != nil {
			return err
		}
	}
	for _, path := range pathsToDelete {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
