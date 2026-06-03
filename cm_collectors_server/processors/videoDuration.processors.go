package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	processorsffmpeg "cm_collectors_server/processorsFFmpeg"
	"cm_collectors_server/utils"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type VideoDuration struct{}

// videoDurationProbeQueue 是视频时长采集的全局后台队列。
//
// 这里故意不使用“每个分集一个 goroutine”的并发模式，因为列表页、编辑保存、批量资源
// 同时触发时，短时间内可能产生大量 ffprobe 和数据库更新。并发过高会让数据库写锁竞争变
// 得明显，也可能让前端感受到卡顿。
//
// 队列的策略是：
// 1. 同一时间只允许一个后台 worker 顺序处理任务；
// 2. ids 记录“已排队或正在执行”的分集 ID，防止同一个分集被重复加入队列；
// 3. 每处理完一个分集后间隔一小段时间，再继续处理下一个，让数据库和磁盘都有喘息空间。
var videoDurationProbeQueue = struct {
	sync.Mutex
	ids     map[string]bool
	items   []models.ResourcesDramaSeries
	running bool
}{
	ids: map[string]bool{},
}

// TriggerForResources 按资源列表触发视频时长采集。
//
// 这个方法通常由资源列表接口调用。它只负责把当前页中“缺失时长且需要采集”的视频分集
// 放入后台队列，不等待 ffprobe 完成，因此不会阻塞列表接口返回。
func (VideoDuration) TriggerForResources(filesBasesID string, resources *[]models.Resources) {
	if resources == nil || !videoDurationEnabled(filesBasesID) {
		return
	}

	var needProbe []models.ResourcesDramaSeries
	for _, resource := range *resources {
		if !isVideoResourceMode(resource.Mode) {
			continue
		}
		for _, ds := range resource.ResourcesDramaSeries {
			if shouldProbeDuration(ds) {
				needProbe = append(needProbe, ds)
			}
		}
	}

	VideoDuration{}.TriggerForDramaSeries(needProbe)
}

// TriggerForResource 按单个资源触发视频时长采集。
//
// 这个方法主要用于新建资源、修改资源成功之后。它同样只入队，不等待采集结果，避免保存
// 操作因为 ffprobe 或磁盘读取变慢。
func (VideoDuration) TriggerForResource(resource *models.Resources) {
	if resource == nil || !videoDurationEnabled(resource.FilesBasesID) || !isVideoResourceMode(resource.Mode) {
		return
	}

	var needProbe []models.ResourcesDramaSeries
	for _, ds := range resource.ResourcesDramaSeries {
		if shouldProbeDuration(ds) {
			needProbe = append(needProbe, ds)
		}
	}

	VideoDuration{}.TriggerForDramaSeries(needProbe)
}

// TriggerForDramaSeries 把需要采集时长的分集加入后台队列。
//
// 这里只做轻量判断和入队：如果同一个分集 ID 已经在队列里或正在采集中，就直接跳过。
// 入队完成后启动单 worker；如果 worker 已经在运行，startDurationProbeWorker 会直接返回。
func (VideoDuration) TriggerForDramaSeries(dramaSeries []models.ResourcesDramaSeries) {
	hasNewJob := false
	for _, ds := range dramaSeries {
		if enqueueDurationProbe(ds) {
			hasNewJob = true
		}
	}

	if hasNewJob {
		startDurationProbeWorker()
	}
}

// runQueue 是唯一的后台消费循环。
//
// 它每次只取出一个分集执行 ffprobe 和数据库更新，执行完成后释放该分集 ID 的占用状态。
// 处理完一个任务后固定等待 1.5 秒，降低连续写数据库、连续读取视频元数据带来的压力。
func (VideoDuration) runQueue() {
	for {
		ds, ok := popDurationProbe()
		if !ok {
			return
		}

		VideoDuration{}.probe(ds)
		releaseDurationProbe(ds.ID)

		time.Sleep(1500 * time.Millisecond)
	}
}

// probe 对单个分集执行真实的视频时长采集，并把结果写回数据库。
//
// 成功时写入 durationSeconds、durationProbeStatus=success、durationProbeTime。
// 失败时写入 durationProbeStatus=failed、durationProbeTime，避免列表页每次打开都反复尝试
// 同一个不可采集的视频。失败不会影响资源保存和列表展示。
func (VideoDuration) probe(ds models.ResourcesDramaSeries) {
	now := datatype.CustomTime(core.TimeNow())
	if ds.Src == "" || !utils.FileExists(ds.Src) {
		models.ResourcesDramaSeries{}.UpdateDuration(core.DBS(), ds.ID, 0, models.DurationProbeStatusFailed, &now)
		return
	}

	basicInfo, err := processorsffmpeg.VideoInfo{}.GetVideoBasicInfo(ds.Src)
	if err != nil {
		core.LogErr(fmt.Errorf("获取视频时长失败：%s，%w", ds.Src, err))
		models.ResourcesDramaSeries{}.UpdateDuration(core.DBS(), ds.ID, 0, models.DurationProbeStatusFailed, &now)
		return
	}

	durationFloat, err := strconv.ParseFloat(basicInfo.Duration, 64)
	if err != nil || durationFloat <= 0 {
		if err != nil {
			core.LogErr(fmt.Errorf("解析视频时长失败：%s，duration=%s，%w", ds.Src, basicInfo.Duration, err))
		}
		models.ResourcesDramaSeries{}.UpdateDuration(core.DBS(), ds.ID, 0, models.DurationProbeStatusFailed, &now)
		return
	}

	models.ResourcesDramaSeries{}.UpdateDuration(core.DBS(), ds.ID, int(durationFloat), models.DurationProbeStatusSuccess, &now)
}

// videoDurationEnabled 读取文件库配置，判断当前文件库是否允许显示和自动采集视频时长。
//
// 配置关闭时，列表、新建、修改都不会触发后台采集；这样旧用户默认不会产生额外的磁盘读取
// 和数据库写入。
func videoDurationEnabled(filesBasesID string) bool {
	config, err := FilesBases{}.Config_FilesBases(filesBasesID)
	if err != nil {
		core.LogErr(fmt.Errorf("读取文件库视频时长配置失败：%s，%w", filesBasesID, err))
		return false
	}

	return config.ShowVideoDuration
}

// isVideoResourceMode 判断资源类型是否属于视频资源。
//
// 当前系统里视频资源主要包括普通视频和短视频；非视频资源即使历史分集里已有时长字段，
// 也不触发采集、不在前端显示。
func isVideoResourceMode(mode datatype.E_resourceMode) bool {
	return mode == datatype.E_resourceMode_Movies || mode == datatype.E_resourceMode_VideoLink
}

// shouldProbeDuration 判断某个分集是否还需要采集时长。
//
// 已有有效时长的分集不重复采集；已经失败过的分集也不在列表浏览时反复尝试，避免不存在
// 的文件或 ffprobe 不可用时持续制造后台任务。
func shouldProbeDuration(ds models.ResourcesDramaSeries) bool {
	// 没有 ID 的分集无法回写数据库；已经有有效时长或成功状态的分集，不再重复采集。
	if ds.ID == "" || ds.DurationSeconds > 0 || ds.DurationProbeStatus == models.DurationProbeStatusSuccess {
		return false
	}

	// failed 不是永久放弃，而是进入冷却期：未满 24 小时跳过，超过 24 小时允许重新入队重试。
	if ds.DurationProbeStatus == models.DurationProbeStatusFailed {
		return durationProbeFailedExpired(ds.DurationProbeTime)
	}

	// 空状态、旧数据或其它未知状态，只要没有有效时长，都允许采集。
	return true
}

// durationProbeFailedExpired 判断失败记录是否已经超过 24 小时重试冷却期。
//
// 这样可以兼顾两件事：
// 1. 文件不存在、ffprobe 不可用等失败不会在每次打开列表时反复刷数据库；
// 2. 网络盘稍后挂载、文件稍后恢复、ffprobe 稍后修好时，系统仍然有机会自动补齐时长。
func durationProbeFailedExpired(probeTime *datatype.CustomTime) bool {
	if probeTime == nil || probeTime.IsZero() {
		return true
	}

	return core.TimeNow().Sub(time.Time(*probeTime)) >= 24*time.Hour
}

// enqueueDurationProbe 尝试把分集加入采集队列。
//
// 返回 true 表示本次确实新增了任务；返回 false 表示分集 ID 无效，或者该分集已经在队列
// 中/正在采集中。
func enqueueDurationProbe(ds models.ResourcesDramaSeries) bool {
	if ds.ID == "" {
		return false
	}

	videoDurationProbeQueue.Lock()
	defer videoDurationProbeQueue.Unlock()

	if videoDurationProbeQueue.ids[ds.ID] {
		return false
	}

	videoDurationProbeQueue.ids[ds.ID] = true
	videoDurationProbeQueue.items = append(videoDurationProbeQueue.items, ds)
	return true
}

// startDurationProbeWorker 启动唯一的后台队列 worker。
//
// 如果 worker 已经在运行，只要把任务加入队列即可；运行中的 worker 会继续把新任务取完。
// 如果 worker 未运行，则标记 running 并启动一个 goroutine 顺序消费队列。
func startDurationProbeWorker() {
	videoDurationProbeQueue.Lock()
	if videoDurationProbeQueue.running {
		videoDurationProbeQueue.Unlock()
		return
	}

	videoDurationProbeQueue.running = true
	videoDurationProbeQueue.Unlock()

	go VideoDuration{}.runQueue()
}

// popDurationProbe 从队列头部取出一个待采集分集。
//
// 队列为空时会把 running 置回 false，让后续新任务可以重新启动 worker。
func popDurationProbe() (models.ResourcesDramaSeries, bool) {
	videoDurationProbeQueue.Lock()
	defer videoDurationProbeQueue.Unlock()

	if len(videoDurationProbeQueue.items) == 0 {
		videoDurationProbeQueue.running = false
		return models.ResourcesDramaSeries{}, false
	}

	ds := videoDurationProbeQueue.items[0]
	videoDurationProbeQueue.items = videoDurationProbeQueue.items[1:]
	return ds, true
}

// releaseDurationProbe 释放分集 ID 的排队/执行占用状态。
//
// 释放后，如果用户再次修改了这个分集路径，或者后续逻辑认为它还需要采集，才可以重新入队。
func releaseDurationProbe(id string) {
	if id == "" {
		return
	}

	videoDurationProbeQueue.Lock()
	delete(videoDurationProbeQueue.ids, id)
	videoDurationProbeQueue.Unlock()
}
