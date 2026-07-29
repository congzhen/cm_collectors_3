package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"errors"

	"gorm.io/gorm"
)

type CronJobs struct{}

func (CronJobs) DataList() (*[]models.CronJobs, error) {
	list, err := models.CronJobs{}.DataList(core.DBS())
	if err != nil {
		return nil, err
	}
	for i := range *list {
		(*list)[i].Running = cronJobExecutions.isRunning((*list)[i].ID)
	}
	return list, nil
}

func (CronJobs) InfoByID_DB(db *gorm.DB, id string) (*models.CronJobs, error) {
	info, err := models.CronJobs{}.Info(db, id)
	if err == nil && info.ID == "" || err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_CronJobs_Not_Found
		return info, err
	}
	return info, nil
}

func (t CronJobs) Exec(id string) error {
	info, err := t.InfoByID_DB(core.DBS(), id)
	if err != nil {
		return err
	}
	return CronJobsExec{}.ExecuteJob(*info)
}

func (t CronJobs) Create(filesBasesID string, jobsType datatype.E_cronJobsType, cronExpression string) (*models.CronJobs, error) {
	return t.CreateScoped(datatype.ReqParam_CreateCronJobs{
		FilesBasesId: filesBasesID, JobsType: jobsType, CronExpression: cronExpression,
	})
}

func (t CronJobs) CreateScoped(par datatype.ReqParam_CreateCronJobs) (*models.CronJobs, error) {
	normalizeCronJobScope(&par)
	if err := validateCronJobScopeSelection(par.ScopeMode, par.FilesBasesIds, par.JobsType); err != nil {
		return nil, err
	}
	db := core.DBS()
	id := core.GenerateUniqueID()
	createdAt := datatype.CustomTime(core.TimeNow())
	cronJobsModels := models.CronJobs{
		ID:             id,
		FilesBasesId:   firstString(par.FilesBasesIds),
		ScopeMode:      par.ScopeMode,
		ConfigJsonData: par.ConfigJsonData,
		JobsType:       par.JobsType,
		CronExpression: par.CronExpression,
		CreatedAt:      &createdAt,
		Status:         true,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := cronJobsModels.Create(tx, &cronJobsModels); err != nil {
			return err
		}
		return (models.CronJobsFilesBases{}).Replace(tx, id, par.FilesBasesIds)
	})
	if err != nil {
		return nil, err
	}
	RestartCronjob()
	return t.InfoByID_DB(db, id)
}
func (t CronJobs) UpdateExec(id string, execError error) error {
	db := core.DBS()
	lastExecStatus := true
	lastExecError := ""
	lastExecAt := datatype.CustomTime(core.TimeNow())
	if execError != nil {
		lastExecError = execError.Error()
		lastExecStatus = false
	}
	cronJobsModels := models.CronJobs{
		ID:             id,
		LastExecError:  lastExecError,
		LastExecStatus: lastExecStatus,
		LastExecAt:     &lastExecAt,
	}
	return cronJobsModels.Update(db, &cronJobsModels, []string{"last_exec_error", "last_exec_status", "last_exec_at"})
}
func (t CronJobs) Update(id string, filesBasesID string, jobsType datatype.E_cronJobsType, cronExpression string) (*models.CronJobs, error) {
	return t.UpdateScoped(id, datatype.ReqParam_CreateCronJobs{
		FilesBasesId: filesBasesID, JobsType: jobsType, CronExpression: cronExpression,
	})
}

func (t CronJobs) UpdateScoped(id string, par datatype.ReqParam_CreateCronJobs) (*models.CronJobs, error) {
	normalizeCronJobScope(&par)
	if err := validateCronJobScopeSelection(par.ScopeMode, par.FilesBasesIds, par.JobsType); err != nil {
		return nil, err
	}
	db := core.DBS()
	cronJobsModels := models.CronJobs{
		ID:             id,
		FilesBasesId:   firstString(par.FilesBasesIds),
		ScopeMode:      par.ScopeMode,
		ConfigJsonData: par.ConfigJsonData,
		JobsType:       par.JobsType,
		CronExpression: par.CronExpression,
		Status:         true,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := cronJobsModels.Update(tx, &cronJobsModels,
			[]string{"filesBases_id", "scope_mode", "config_json_data", "jobs_type", "cron_expression"}); err != nil {
			return err
		}
		return (models.CronJobsFilesBases{}).Replace(tx, id, par.FilesBasesIds)
	})
	if err != nil {
		return nil, err
	}
	RestartCronjob()
	return t.InfoByID_DB(db, id)
}

func cronJobRequiresFilesBase(jobsType datatype.E_cronJobsType) bool {
	return jobsType != datatype.E_cronJobsType_ClearPerformerAvatarCache
}

func normalizeCronJobScope(par *datatype.ReqParam_CreateCronJobs) {
	if len(par.FilesBasesIds) == 0 && par.FilesBasesId != "" {
		par.FilesBasesIds = []string{par.FilesBasesId}
	}
	par.FilesBasesIds = uniqueVideoMetadataStrings(par.FilesBasesIds)
	if par.ScopeMode != models.VideoMetadataScopeAll {
		par.ScopeMode = models.VideoMetadataScopeSelected
	}
	if !cronJobRequiresFilesBase(par.JobsType) {
		par.ScopeMode = models.VideoMetadataScopeAll
		par.FilesBasesIds = nil
	}
}

func validateCronJobScopeSelection(scopeMode string, filesBasesIDs []string, jobsType datatype.E_cronJobsType) error {
	if !cronJobRequiresFilesBase(jobsType) {
		return nil
	}
	if jobsType == datatype.E_cronJobsType_VideoMetadata {
		return validateVideoMetadataScope(scopeMode, filesBasesIDs)
	}
	if scopeMode != models.VideoMetadataScopeSelected || len(filesBasesIDs) != 1 {
		return errors.New("该计划任务必须选择一个执行文件库")
	}
	return nil
}

// validateCronJobScope 保留旧的单库校验入口，供现有任务和测试继续使用。
func validateCronJobScope(filesBasesID string, jobsType datatype.E_cronJobsType) error {
	ids := []string{}
	if filesBasesID != "" {
		ids = append(ids, filesBasesID)
	}
	return validateCronJobScopeSelection(models.VideoMetadataScopeSelected, ids, jobsType)
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (CronJobs) Delete(id string) error {
	err := models.CronJobs{}.DeleteById(core.DBS(), id)
	if err != nil {
		return err
	}
	RestartCronjob()
	return nil
}
func (CronJobs) DeleteByFilesBasesID(db *gorm.DB, filesBasesID string) error {
	err := models.CronJobs{}.DeleteByFilesBasesID(db, filesBasesID)
	if err != nil {
		return err
	}
	RestartCronjob()
	return nil
}
