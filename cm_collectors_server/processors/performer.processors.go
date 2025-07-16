package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"fmt"
	"os"
	"path"
)

type Performer struct{}

func (Performer) BasicList(performerBasesIds []string, careerPerformer, careerDirector bool) (*[]models.PerformerBasic, error) {
	return models.Performer{}.BasicList(core.DBS(), performerBasesIds, careerPerformer, careerDirector)
}

func (Performer) DataList(performerBasesId string, fetchCount bool, page, limit int, search, star, cup string) (*[]models.Performer, int64, error) {
	return models.Performer{}.DataList(core.DBS(), performerBasesId, fetchCount, page, limit, search, star, cup)
}

func (Performer) ListTopPreferredPerformers(preferredIds []string, mainPerformerBasesId string, shieldNoPerformerPhoto bool, limit int) (*[]models.Performer, error) {
	return models.Performer{}.ListTopPreferredPerformers(core.DBS(), preferredIds, mainPerformerBasesId, shieldNoPerformerPhoto, limit)
}

func (Performer) RecycleBin(performerBasesId string) (*[]models.Performer, error) {
	return models.Performer{}.RecycleBin(core.DBS(), performerBasesId)
}

// SavePerformerPhoto 保存或更新表演者的图片
func (Performer) SavePerformerPhoto(par *datatype.ReqParam_PerformerData) (string, error) {
	if par.PhotoBase64 == "" || par.Performer.PerformerBasesID == "" {
		return "", nil // 如果没有图片数据或PerformerBasesID，则不处理
	}

	// 生成唯一的图片名称
	photoName := fmt.Sprintf("%s.jpg", core.GenerateUniqueID())
	filePath := path.Join(core.Config.System.FilePath, "performerFace", par.Performer.PerformerBasesID, photoName)

	// 将Base64图片保存为文件
	err := utils.SaveBase64AsImage(par.PhotoBase64, filePath)
	if err != nil {
		return "", errorMessage.WrapError(errorMessage.Err_performer_Save_Photo_Failed, err)
	}
	return photoName, nil
}

// DeletePerformerPhoto 删除旧的表演者图片
func (Performer) DeletePerformerPhoto(performerBasesID, photoName string) error {
	if photoName == "" {
		return nil // 如果没有旧图片，则不处理
	}
	oldFilePath := path.Join(core.Config.System.FilePath, "performerFace", performerBasesID, photoName)
	err := os.Remove(oldFilePath)
	if err != nil && !os.IsNotExist(err) {
		return errorMessage.WrapError(errorMessage.Err_performer_Delete_Photo_Failed, err)
	}
	return nil
}

// Create 创建新的表演者数据
func (t Performer) Create(par *datatype.ReqParam_PerformerData) (*models.Performer, error) {
	// 保存图片
	photoName, err := t.SavePerformerPhoto(par)
	if err != nil {
		return nil, err
	}

	db := core.DBS()
	createdAt := datatype.CustomTime(core.TimeNow())
	id := core.GenerateUniqueID()

	performerModels := models.Performer{
		ID:               id,
		PerformerBasesID: par.Performer.PerformerBasesID,
		Name:             par.Performer.Name,
		AliasName:        par.Performer.AliasName,
		KeyWords:         utils.PinyinInitials(par.Performer.Name + par.Performer.AliasName),
		Birthday:         par.Performer.Birthday,
		Nationality:      par.Performer.Nationality,
		CareerPerformer:  par.Performer.CareerPerformer,
		CareerDirector:   par.Performer.CareerDirector,
		Photo:            photoName,
		Introduction:     par.Performer.Introduction,
		Cup:              par.Performer.Cup,
		Bust:             par.Performer.Bust,
		Hip:              par.Performer.Hip,
		Waist:            par.Performer.Waist,
		Stars:            par.Performer.Stars,
		RetreatStatus:    par.Performer.RetreatStatus,
		CreatedAt:        &createdAt,
		Status:           par.Performer.Status,
	}

	if err := performerModels.Create(db, &performerModels); err != nil {
		return nil, err
	}
	return &performerModels, nil
}

func (t Performer) UpdatePerformerStatus(id string, status bool) error {
	return models.Performer{}.Update(core.DBS(), &models.Performer{ID: id, Status: status}, []string{"status"})
}

// Update 更新表演者数据
func (t Performer) Update(par *datatype.ReqParam_PerformerData) (*models.Performer, error) {
	// 保存新图片
	newPhotoName, err := t.SavePerformerPhoto(par)
	if err != nil {
		return nil, err
	}

	// 删除旧图片
	if newPhotoName != "" {
		t.DeletePerformerPhoto(par.Performer.PerformerBasesID, par.Performer.Photo)
	}

	db := core.DBS()
	performerModels := models.Performer{
		ID:               par.Performer.ID,
		PerformerBasesID: par.Performer.PerformerBasesID,
		Name:             par.Performer.Name,
		AliasName:        par.Performer.AliasName,
		KeyWords:         utils.PinyinInitials(par.Performer.Name + par.Performer.AliasName),
		Birthday:         par.Performer.Birthday,
		Nationality:      par.Performer.Nationality,
		CareerPerformer:  par.Performer.CareerPerformer,
		CareerDirector:   par.Performer.CareerDirector,
		Introduction:     par.Performer.Introduction,
		Cup:              par.Performer.Cup,
		Bust:             par.Performer.Bust,
		Hip:              par.Performer.Hip,
		Waist:            par.Performer.Waist,
		Stars:            par.Performer.Stars,
		RetreatStatus:    par.Performer.RetreatStatus,
		Status:           par.Performer.Status,
	}
	fieldsToUpdate := []string{
		"performerBases_id",
		"name",
		"aliasName",
		"keyWords",
		"birthday",
		"nationality",
		"careerDirector",
		"careerPerformer",
		"introduction",
		"cup",
		"bust",
		"waist",
		"hip",
		"stars",
		"retreatStatus",
		"status",
	}

	// 更新图片字段
	if newPhotoName != "" {
		performerModels.Photo = newPhotoName
		fieldsToUpdate = append(fieldsToUpdate, "photo")
	}

	if err := performerModels.Update(db, &performerModels, fieldsToUpdate); err != nil {
		return nil, err
	}
	return &performerModels, nil
}
