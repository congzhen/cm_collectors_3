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
	if err := validateCronJobScope(filesBasesID, jobsType); err != nil {
		return nil, err
	}
	if !cronJobRequiresFilesBase(jobsType) {
		filesBasesID = ""
	}
	db := core.DBS()
	id := core.GenerateUniqueID()
	createdAt := datatype.CustomTime(core.TimeNow())
	cronJobsModels := models.CronJobs{
		ID:             id,
		FilesBasesId:   filesBasesID,
		JobsType:       jobsType,
		CronExpression: cronExpression,
		CreatedAt:      &createdAt,
		Status:         true,
	}
	err := cronJobsModels.Create(db, &cronJobsModels)
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
	if err := validateCronJobScope(filesBasesID, jobsType); err != nil {
		return nil, err
	}
	if !cronJobRequiresFilesBase(jobsType) {
		filesBasesID = ""
	}
	db := core.DBS()
	cronJobsModels := models.CronJobs{
		ID:             id,
		FilesBasesId:   filesBasesID,
		JobsType:       jobsType,
		CronExpression: cronExpression,
		Status:         true,
	}
	err := cronJobsModels.Update(db, &cronJobsModels, []string{"filesBases_id", "jobs_type", "cron_expression"})
	if err != nil {
		return nil, err
	}
	RestartCronjob()
	return t.InfoByID_DB(db, id)
}

func cronJobRequiresFilesBase(jobsType datatype.E_cronJobsType) bool {
	return jobsType != datatype.E_cronJobsType_ClearPerformerAvatarCache
}

func validateCronJobScope(filesBasesID string, jobsType datatype.E_cronJobsType) error {
	if cronJobRequiresFilesBase(jobsType) && filesBasesID == "" {
		return errors.New("该计划任务必须选择执行文件库")
	}
	return nil
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
