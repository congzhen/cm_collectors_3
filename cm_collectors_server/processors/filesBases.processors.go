package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"

	"gorm.io/gorm"
)

type FilesBases struct{}

func (FilesBases) DataList() (*[]models.FilesBases, error) {
	return models.FilesBases{}.DataList(core.DBS())
}

func (FilesBases) InfoById(id string) (*models.FilesBases, error) {
	return models.FilesBases{}.Info(core.DBS(), id)
}

func (FilesBases) InfoDetailsById(id string) (*models.FilesBasesDetails, error) {
	return models.FilesBases{}.InfoDetails(core.DBS(), id)
}

// 设置FilesBases信息
func (t FilesBases) SetFilesBases(par *datatype.ReqParam_SetFilesBases) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := t.updateFilesBasesInfo(tx, par); err != nil {
			return err
		}
		if err := t.updateFilesBasesSetting(tx, par.ID, par.Config); err != nil {
			return err
		}
		if err := t.syncPerformerBasesRelations(tx, par); err != nil {
			return err
		}
		return nil
	})
}

// 更新FilesBases信息
func (FilesBases) updateFilesBasesInfo(tx *gorm.DB, par *datatype.ReqParam_SetFilesBases) error {
	filesBasesModel := models.FilesBases{
		ID:     par.ID,
		Name:   par.Info.Name,
		Sort:   par.Info.Sort,
		Status: par.Info.Status,
	}
	return filesBasesModel.Update(tx, &filesBasesModel, []string{"name", "sort", "status"})
}

// 更新config配置信息
func (FilesBases) updateFilesBasesSetting(tx *gorm.DB, filesBasesID, config string) error {
	settingModel := models.FilesBasesSetting{
		ConfigJsonData: config,
	}
	return settingModel.Update(tx, filesBasesID, &settingModel, []string{"config_json_data"})
}

// 同步演员集关联关系
func (t FilesBases) syncPerformerBasesRelations(tx *gorm.DB, par *datatype.ReqParam_SetFilesBases) error {
	existingList, err := models.FilesRelatedPerformerBases{}.LisyByFilesBasesID(tx, par.ID)
	if err != nil {
		return err
	}

	existingMap := make(map[string]models.FilesRelatedPerformerBases)
	existingIDs := make([]string, 0, len(*existingList))

	for _, item := range *existingList {
		existingIDs = append(existingIDs, item.PerformerBasesID)
		existingMap[item.PerformerBasesID] = item
	}

	toUpdate, toDelete, toAdd := utils.ArrayIntersectDiff(existingIDs, par.RelatedPerformerBases)

	if err := t.handleUpdates(tx, existingMap, toUpdate, par.MainPerformerBasesId); err != nil {
		return err
	}
	if err := t.handleDeletes(tx, existingMap, toDelete); err != nil {
		return err
	}
	if err := t.handleAdds(tx, toAdd, par.ID, par.MainPerformerBasesId); err != nil {
		return err
	}

	return nil
}

// 处理需要更新的关联记录
func (FilesBases) handleUpdates(tx *gorm.DB, existingMap map[string]models.FilesRelatedPerformerBases, toUpdate []string, mainPerformerID string) error {
	for _, id := range toUpdate {
		item := existingMap[id]
		isCurrentMain := item.Main
		shouldSetMain := item.PerformerBasesID == mainPerformerID

		if isCurrentMain == shouldSetMain {
			continue
		}

		updateModel := models.FilesRelatedPerformerBases{
			ID:               item.ID,
			Main:             shouldSetMain,
			PerformerBasesID: id,
		}

		if err := updateModel.Update(tx, &updateModel, []string{"main", "performerBases_id"}); err != nil {
			return err
		}
	}
	return nil
}

// 删除不再存在的关联
func (FilesBases) handleDeletes(tx *gorm.DB, existingMap map[string]models.FilesRelatedPerformerBases, toDelete []string) error {
	if len(toDelete) > 0 {
		ids := make([]string, len(toDelete))
		for i, id := range toDelete {
			ids[i] = existingMap[id].ID
		}
		return models.FilesRelatedPerformerBases{}.DeleteIDS(tx, ids)
	}
	return nil
}

// 新增新的关联
func (FilesBases) handleAdds(tx *gorm.DB, toAdd []string, filesBasesID string, mainPerformerID string) error {
	if len(toAdd) == 0 {
		return nil
	}
	var newRecords []models.FilesRelatedPerformerBases
	for _, id := range toAdd {
		newRecords = append(newRecords, models.FilesRelatedPerformerBases{
			ID:               core.GenerateUniqueID(),
			FilesBasesID:     filesBasesID,
			PerformerBasesID: id,
			Main:             id == mainPerformerID,
		})
	}

	return models.FilesRelatedPerformerBases{}.Creates(tx, &newRecords)
}

func (FilesBases) GetTotal() (int64, error) {
	return models.FilesBases{}.GetTotal(core.DBS())
}

func (t FilesBases) Create(name, mainPerformerBasesId string, relatedPerformerBasesIds []string) (string, error) {
	db := core.DBS()
	id := core.GenerateUniqueID()
	tagTotal, err := t.GetTotal()
	if err != nil {
		return id, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		createdAt := datatype.CustomTime(core.TimeNow())
		filesBaseModels := models.FilesBases{
			ID:        id,
			Name:      name,
			Sort:      int(tagTotal) + 1,
			CreatedAt: &createdAt,
			Status:    true,
		}
		err := filesBaseModels.Create(tx, &filesBaseModels)
		if err != nil {
			return err
		}
		err = models.FilesBasesSetting{}.CreateNull(tx, id)
		if err != nil {
			return err
		}
		filesRelatedPerformerBasesModelsSlc := []models.FilesRelatedPerformerBases{}
		if len(relatedPerformerBasesIds) > 0 {
			for _, performerBasesID := range relatedPerformerBasesIds {
				main := false
				if performerBasesID == mainPerformerBasesId {
					main = true
				}
				relatedModels := models.FilesRelatedPerformerBases{
					ID:               core.GenerateUniqueID(),
					FilesBasesID:     id,
					PerformerBasesID: performerBasesID,
					Main:             main,
				}
				filesRelatedPerformerBasesModelsSlc = append(filesRelatedPerformerBasesModelsSlc, relatedModels)
			}
		} else {
			filesRelatedPerformerBasesModelsSlc = append(filesRelatedPerformerBasesModelsSlc, models.FilesRelatedPerformerBases{
				ID:               core.GenerateUniqueID(),
				FilesBasesID:     id,
				PerformerBasesID: mainPerformerBasesId,
				Main:             true,
			})
		}
		err = models.FilesRelatedPerformerBases{}.Creates(tx, &filesRelatedPerformerBasesModelsSlc)
		if err != nil {
			return err
		}
		return nil
	})
	return id, err
}
