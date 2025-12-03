package processors

import (
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// 计划任务
var cronScheduler *cron.Cron

type CronJobsExec struct{}

// InitCronjob 初始化定时任务
// logPath: 定时任务日志的保存路径
// 该函数负责初始化定时任务相关的配置
func InitCronjob() {
	// 创建一个新的cron调度器，并支持秒级调度
	cronScheduler = cron.New(cron.WithSeconds())
	cje := CronJobsExec{}
	err := cje.RegCronJobs()
	if err != nil {
		fmt.Println(err)
		return
	}
	cje.Start()
}

func RestartCronjob() error {
	cje := CronJobsExec{}
	cje.Clear()
	err := cje.RegCronJobs()
	if err != nil {
		fmt.Println(err)
		return err
	}
	cje.Start()
	return nil
}

func (t CronJobsExec) GetCronJobsEntries() []cron.Entry {
	return cronScheduler.Entries()
}

func (t CronJobsExec) RegCronJobs() error {
	list, err := CronJobs{}.DataList()
	if err != nil {
		return err
	}

	for _, job := range *list {
		// 创建局部变量副本以避免闭包问题
		jobCopy := job

		// 将任务添加到调度器
		_, err := cronScheduler.AddFunc(job.CronExpression, func() {
			fmt.Printf("开始执行计划任务: %s (%s)\n", job.FilesBases.Name, job.JobsType)

			// 执行任务
			execErr := t.ExecuteJob(jobCopy)
			if execErr != nil {
				fmt.Printf("计划任务执行完成 [%s]: %v\n", job.JobsType, execErr)
			} else {
				fmt.Printf("计划任务执行完成 [%s]: 成功\n", job.JobsType)
			}
		})

		if err != nil {
			fmt.Printf("注册计划任务失败 [%s] %s: %v\n", job.FilesBases.Name, job.CronExpression, err)
			return fmt.Errorf("注册计划任务 '%s' 失败: %w", job.FilesBases.Name, err)
		}

		fmt.Printf("成功注册计划任务: %s (%s) - %s\n", job.FilesBases.Name, job.JobsType, job.CronExpression)
	}

	return nil
}
func (t CronJobsExec) ExecuteJob(data models.CronJobs) error {
	// 执行任务核心逻辑
	err := t.executeJobTask(data)

	// 更新任务执行结果
	CronJobs{}.UpdateExec(data.ID, err)

	if err != nil {
		fmt.Printf("计划任务执行失败 [%s]: %v\n", data.JobsType, err)
		return err
	}

	fmt.Printf("计划任务执行成功 [%s]\n", data.JobsType)
	return nil
}
func (t CronJobsExec) executeJobTask(data models.CronJobs) error {
	switch data.JobsType {
	case datatype.E_cronJobsType_Clear:
		// 清理任务处理
		return t.cronJobs_Clear(data)
	case datatype.E_cronJobsType_ScraperPerformer:
		// 刮削演员任务处理
		return t.cronJobs_ScraperPerformer(data)
	case datatype.E_cronJobsType_ScraperResource:
		// 刮削资源任务处理
		return t.cronJobs_ScraperResource(data)
	case datatype.E_cronJobsType_Import:
		// 导入任务处理
		return t.cronJobs_Import(data)
	default:
		return fmt.Errorf("未知任务类型: %s", data.JobsType)
	}
}
func (t CronJobsExec) Start() {
	cronScheduler.Start()
}
func (CronJobsExec) Stop() {
	cronScheduler.Stop()
}
func (t CronJobsExec) Clear() {
	// 获取所有已注册的任务
	entries := cronScheduler.Entries()

	// 遍历并移除所有任务
	for _, entry := range entries {
		fmt.Println("移除计划任务:", entry.ID, entry.Job)
		cronScheduler.Remove(entry.ID)
	}
	t.Stop()
}

// 清理任务处理
func (t CronJobsExec) cronJobs_Clear(data models.CronJobs) error {
	fmt.Println("执行计划任务:", data.FilesBases.Name, data.JobsType, data.CronExpression)
	delList, err := Resources{}.DataListDeletedResource([]string{data.FilesBasesId})
	if err != nil {
		return err
	}
	for _, v := range *delList {
		err := Resources{}.DeleteResource(v.ID)
		if err != nil {
			continue
		}
	}
	return nil
}

// 刮削演员任务处理
func (t CronJobsExec) cronJobs_ScraperPerformer(data models.CronJobs) error {
	fmt.Println("执行计划任务:", data.FilesBases.Name, data.JobsType, data.CronExpression)
	config, err := FilesBases{}.Config_ScraperPerformer(data.FilesBasesId)
	if err != nil {
		return err
	}
	filesBasesInfo, err := FilesBases{}.InfoById(data.FilesBasesId)
	if err != nil {
		return err
	}
	for _, v := range *&filesBasesInfo.FilesRelatedPerformerBases {
		performerBasesId := v.PerformerBasesID
		datalist, err := Performer{}.SearchLastScraperUpdateTime(performerBasesId, config.LastScraperUpdateTime)
		if err != nil {
			return err
		}
		for _, d := range *datalist {
			par := datatype.ReqParam_ScraperPerformerDataProcess{
				Operate:          config.Operate,
				PerformerId:      d.ID,
				PerformerName:    d.Name,
				ScraperConfig:    config.ScraperConfig,
				PerformerBasesId: performerBasesId,
			}
			err := Scraper{}.ScraperPerformerDataProcess(&par)
			if err != nil {
				continue
			}
		}
	}
	return nil
}

// 刮削资源任务处理
func (t CronJobsExec) cronJobs_ScraperResource(data models.CronJobs) error {
	fmt.Println("执行计划任务:", data.FilesBases.Name, data.JobsType, data.CronExpression)
	config, err := FilesBases{}.Config_Scraper(data.FilesBasesId)
	if err != nil {
		return err
	}
	pendingFilePaths, err := Scraper{}.Pretreatment(data.FilesBasesId, config, false)
	if err != nil {
		return err
	}
	for _, v := range pendingFilePaths {
		err := Scraper{}.ScraperDataProcess(data.FilesBasesId, v, config)
		if err != nil {
			continue
		}
		// 添加短暂延迟，避免处理过快导致系统负载过高
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// 导入任务处理
func (t CronJobsExec) cronJobs_Import(data models.CronJobs) error {
	fmt.Println("执行计划任务:", data.FilesBases.Name, data.JobsType, data.CronExpression)
	config, err := FilesBases{}.Config_ScanDisk(data.FilesBasesId)
	if err != nil {
		return err
	}
	nonExistingSrcPaths, err := ImportData{}.ScanDiskImportPaths(data.FilesBasesId, config, false)
	if err != nil {
		return err
	}
	for _, v := range nonExistingSrcPaths {
		err := ImportData{}.ScanDiskImportData(data.FilesBasesId, v, config)
		if err != nil {
			continue
		}
		// 添加短暂延迟，避免处理过快导致系统负载过高
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
