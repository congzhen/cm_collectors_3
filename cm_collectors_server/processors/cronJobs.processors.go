package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"

	"gorm.io/gorm"
)

type CronJobs struct{}

func (CronJobs) DataList() (*[]models.CronJobs, error) {
	return models.CronJobs{}.DataList(core.DBS())
}

func (CronJobs) InfoByID_DB(db *gorm.DB, id string) (*models.CronJobs, error) {
	info, err := models.CronJobs{}.Info(db, id)
	if err == nil && info.ID == "" || err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_CronJobs_Not_Found
		return info, err
	}
	return info, nil
}

func (t CronJobs) Create(filesBasesID string, jobsType datatype.E_cronJobsType, cronExpression string) (*models.CronJobs, error) {
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
	return t.InfoByID_DB(db, id)
}
func (t CronJobs) Update(id string, filesBasesID string, jobsType datatype.E_cronJobsType, cronExpression string) (*models.CronJobs, error) {
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
	return t.InfoByID_DB(db, id)
}

func (CronJobs) Delete(id string) error {
	return models.CronJobs{}.DeleteById(core.DBS(), id)
}
