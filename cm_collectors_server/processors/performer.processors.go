package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"sort"

	"gorm.io/gorm"
)

type Performer struct{}

type PerformerExpand struct {
	PerformerInfo models.Performer `json:"performerInfo"`
	PhotoBase64   string           `json:"photoBase64"`
}

func (Performer) BasicList(performerBasesIds []string, careerPerformer, careerDirector bool) (*[]models.PerformerBasic, error) {
	return models.Performer{}.BasicList(core.DBS(), performerBasesIds, careerPerformer, careerDirector)
}

func (Performer) DataList(performerBasesId string, fetchCount bool, page, limit int, search, star, cup, charIndex string) (*[]models.Performer, int64, error) {
	return models.Performer{}.DataList(core.DBS(), performerBasesId, fetchCount, page, limit, search, star, cup, charIndex)
}

// DataListByIds 根据提供的演员ID列表获取演员数据列表
//
// 该函数通过调用模型层的DataListByIds方法从数据库中查询指定ID的演员信息，
// 并可根据ordered参数决定是否按照传入的ID顺序对结果进行排序。
//
// 参数:
//
//	ids []string : 需要查询的演员ID列表
//	ordered bool : 是否按照传入的ID顺序对结果进行排序，true表示需要排序，false表示不需要排序
//
// 返回值:
//
//	*[]models.Performer : 查询到的演员数据列表指针
//	error : 查询过程中发生的错误信息，如果没有错误则为nil
func (Performer) DataListByIds(ids []string, ordered bool) (*[]models.Performer, error) {
	dataList, err := models.Performer{}.DataListByIds(core.DBS(), ids)
	if err != nil {
		return nil, err
	}
	if !ordered {
		return dataList, nil
	}
	//根据ids的顺序排序
	idToIndex := make(map[string]int)
	for i, id := range ids {
		idToIndex[id] = i
	}

	sort.Slice(*dataList, func(i, j int) bool {
		return idToIndex[(*dataList)[i].ID] < idToIndex[(*dataList)[j].ID]
	})

	return dataList, nil
}

func (Performer) ListTopPreferredPerformers(preferredIds []string, mainPerformerBasesId string, shieldNoPerformerPhoto bool, limit int) (*[]models.Performer, error) {
	return models.Performer{}.ListTopPreferredPerformers(core.DBS(), preferredIds, mainPerformerBasesId, shieldNoPerformerPhoto, limit)
}
func (Performer) SearchLastScraperUpdateTime(performerBasesId, lastScraperUpdateTime string) (*[]models.PerformerBasic, error) {
	return models.Performer{}.SearchLastScraperUpdateTime(core.DBS(), performerBasesId, lastScraperUpdateTime)
}
func (Performer) PhotosByPerformerBasesId_DB(db *gorm.DB, performerBasesId string) ([]string, error) {
	return models.Performer{}.PhotosByPerformerBasesId_DB(db, performerBasesId)
}

func (Performer) InfoByName(performerBasesID, name string, searchAliasName bool) (*models.Performer, error) {
	info, err := models.Performer{}.InfoByName(core.DBS(), performerBasesID, name, searchAliasName)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_performer_Not_Found
	}
	return info, err
}
func (t Performer) InfoByID(id string) (*models.Performer, error) {
	return t.InfoByID_DB(core.DBS(), id)
}
func (Performer) InfoByID_DB(db *gorm.DB, id string) (*models.Performer, error) {
	info, err := models.Performer{}.InfoByID(db, id)
	if err != nil && err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_performer_Not_Found
	}
	return info, err
}
func (t Performer) PerformerInfoByNameNotFoundCreate(filesBasesId, name, photoBase64 string) (*models.Performer, error) {
	mainPerformerBasesId, err := FilesBases{}.GetMainPerformerBasesId(filesBasesId)
	if err != nil {
		return nil, err
	}
	info, err := t.InfoByName(mainPerformerBasesId, name, true)
	if err != nil && err == errorMessage.Err_performer_Not_Found {
		par := datatype.ReqParam_PerformerData{
			Performer: datatype.ReqParam_Performer{
				PerformerBasesID: mainPerformerBasesId,
				Name:             name,
				CareerPerformer:  true,
			},
		}
		if photoBase64 != "" {
			par.PhotoBase64 = photoBase64
		}
		return t.Create(&par)
	}
	return info, err
}

func (Performer) RecycleBin(performerBasesId string) (*[]models.Performer, error) {
	return models.Performer{}.RecycleBin(core.DBS(), performerBasesId)
}

// 判断演员图片是否存在
func (Performer) PerformerPhotoExists(performerBasesID, photoName string) bool {
	filePath := path.Join(core.Config.System.FilePath, "performerFace", performerBasesID, photoName)
	_, err := os.Stat(filePath)
	return err == nil
}

// getPerformerPhoto 获取演员图片 base64
func (Performer) GetPerformerPhoto(performerBasesID, photoName string) (string, error) {
	if photoName == "" {
		return "", nil // 如果没有图片名称，则返回空
	}
	filePath := path.Join(core.Config.System.FilePath, "performerFace", performerBasesID, photoName)
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return "", err
	}
	// 读取文件内容
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	buffer := make([]byte, fileInfo.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(buffer)
	return "data:image/jpeg;base64," + base64Str, nil
}

// SavePerformerPhoto 保存或更新表演者的图片
func (t Performer) SavePerformerPhoto(par *datatype.ReqParam_PerformerData) (string, error) {
	return t._savePerformerPhoto(par.Performer.PerformerBasesID, par.PhotoBase64)
}
func (Performer) _savePerformerPhoto(performerBasesID, photoBase64 string) (string, error) {
	if photoBase64 == "" || performerBasesID == "" {
		return "", nil // 如果没有图片数据或PerformerBasesID，则不处理
	}
	// 生成唯一的图片名称
	photoName := fmt.Sprintf("%s.jpg", core.GenerateUniqueID())
	filePath := path.Join(core.Config.System.FilePath, "performerFace", performerBasesID, photoName)
	// 将Base64图片保存为文件
	err := utils.SaveBase64AsImage(photoBase64, filePath, true)
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

func (t Performer) CreateByModelsPerformer_DB(db *gorm.DB, info *models.Performer, PhotoBase64 string) error {
	// 保存图片
	photoName, err := t._savePerformerPhoto(info.PerformerBasesID, PhotoBase64)
	if err != nil {
		return err
	}
	createdAt := datatype.CustomTime(core.TimeNow())
	info.CreatedAt = &createdAt
	info.Photo = photoName
	err = info.Create(db, info)
	if err != nil {
		return err
	}
	return nil
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

func (t Performer) CreateScraperByModels(id string, dataModels models.Performer, perforomerPhotoBase64 string) error {
	if perforomerPhotoBase64 != "" {
		// 保存图片
		photoName, err := t._savePerformerPhoto(dataModels.PerformerBasesID, perforomerPhotoBase64)
		if err != nil {
			return err
		}
		dataModels.Photo = photoName
	}
	db := core.DBS()
	createdAt := datatype.CustomTime(core.TimeNow())
	dataModels.ID = id
	dataModels.KeyWords = utils.PinyinInitials(dataModels.Name + dataModels.AliasName)
	dataModels.CreatedAt = &createdAt
	dataModels.Status = true
	lastScraperUpdateTime := datatype.CustomDate{}
	lastScraperUpdateTime.SetValue(core.TimeNow())
	dataModels.LastScraperUpdateTime = &lastScraperUpdateTime
	return dataModels.Create(db, &dataModels)
}

// UpdateScraperByModels 根据抓取的数据更新演员信息
//
// 该函数支持两种更新模式：
// 1. 更新模式(E_PerformerUpdateOperate_Update)：仅当原字段为空时才更新
// 2. 覆盖模式(E_PerformerUpdateOperate_Cover)：强制更新所有字段
//
// 参数:
//   - id: 演员ID
//   - dataModels: 包含新演员数据的模型
//   - perforomerPhotoBase64: 演员照片的base64编码字符串
//   - updateOperate: 更新操作类型，决定更新策略
//
// 返回值:
//   - error: 更新过程中发生的错误，如果没有错误则为nil
func (t Performer) UpdateScraperByModels(id string, dataModels models.Performer, perforomerPhotoBase64 string, updateOperate datatype.E_ScraperOperate) error {
	dataModels.ID = id
	info, err := t.InfoByID(id)
	if err != nil {
		return err
	}
	dataModels.KeyWords = utils.PinyinInitials(dataModels.Name + dataModels.AliasName)
	var fieldsToUpdate []string

	// 定义所有可能需要更新的字段
	allFields := []struct {
		fieldName string
		isEmpty   func(models.Performer) bool
	}{
		{"name", func(p models.Performer) bool { return p.Name == "" }},
		{"aliasName", func(p models.Performer) bool { return p.AliasName == "" }},
		{"birthday", func(p models.Performer) bool { return p.Birthday == "" }},
		{"nationality", func(p models.Performer) bool { return p.Nationality == "" }},
		{"introduction", func(p models.Performer) bool { return p.Introduction == "" }},
		{"cup", func(p models.Performer) bool { return p.Cup == "" }},
		{"bust", func(p models.Performer) bool { return p.Bust == "" }},
		{"hip", func(p models.Performer) bool { return p.Hip == "" }},
		{"waist", func(p models.Performer) bool { return p.Waist == "" }},
	}

	if updateOperate == datatype.E_PerformerUpdateOperate_Update {
		// 在更新模式下，只更新空白字段
		updateKeyWords := false
		for _, f := range allFields {
			// 检查现有记录中的字段是否为空
			switch f.fieldName {
			case "name":
				if info.Name == "" && dataModels.Name != "" {
					fieldsToUpdate = append(fieldsToUpdate, f.fieldName)
					updateKeyWords = true
				}
			case "aliasName":
				if info.AliasName == "" && dataModels.AliasName != "" {
					fieldsToUpdate = append(fieldsToUpdate, f.fieldName)
					updateKeyWords = true
				}
			default:
				// 对于其他字段，只在当前值为空且新值非空时更新
				if f.isEmpty(*info) {
					fieldsToUpdate = append(fieldsToUpdate, f.fieldName)
				}
			}
		}

		if updateKeyWords {
			fieldsToUpdate = append(fieldsToUpdate, "keyWords")
		}

		// 处理照片更新
		if perforomerPhotoBase64 != "" {
			needToUpdatePhoto := false
			if info.Photo == "" {
				// 如果当前没有照片，直接更新
				needToUpdatePhoto = true
			} else if !t.PerformerPhotoExists(info.PerformerBasesID, info.Photo) {
				// 如果当前有照片但文件不存在，也更新
				needToUpdatePhoto = true
			}

			if needToUpdatePhoto {
				photoID, err := t._savePerformerPhoto(info.PerformerBasesID, perforomerPhotoBase64)
				if err == nil { // 保存成功时才更新photo字段
					dataModels.Photo = photoID
					fieldsToUpdate = append(fieldsToUpdate, "photo")
				}
			}
		}
	} else {
		// 在替换模式下，更新所有字段
		for _, f := range allFields {
			fieldsToUpdate = append(fieldsToUpdate, f.fieldName)
		}
		fieldsToUpdate = append(fieldsToUpdate, "keyWords")

		// 处理照片更新
		if perforomerPhotoBase64 != "" {
			// 在替换模式下，如果已有照片且文件存在，则先删除旧照片
			if info.Photo != "" && t.PerformerPhotoExists(info.PerformerBasesID, info.Photo) {
				t.DeletePerformerPhoto(info.PerformerBasesID, info.Photo)
			}

			// 保存新照片
			photoID, err := t._savePerformerPhoto(info.PerformerBasesID, perforomerPhotoBase64)
			if err == nil { // 保存成功时才更新photo字段
				dataModels.Photo = photoID
				fieldsToUpdate = append(fieldsToUpdate, "photo")
			}
		}
	}
	lastScraperUpdateTime := datatype.CustomDate{}
	lastScraperUpdateTime.SetValue(core.TimeNow())
	dataModels.LastScraperUpdateTime = &lastScraperUpdateTime
	fieldsToUpdate = append(fieldsToUpdate, "lastScraperUpdateTime")
	return dataModels.Update(core.DBS(), &dataModels, fieldsToUpdate)
}

func (t Performer) DeleteByPerformerBasesIds(db *gorm.DB, performerBasesIds []string) error {
	return models.Performer{}.DeleteByPerformerBasesIds(db, performerBasesIds)
}

func (t Performer) DeleteByID(id string) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		info, err := t.InfoByID_DB(tx, id)
		if err != nil {
			return err
		}
		//删除演员
		err = models.Performer{}.DeleteById(tx, id)
		if err != nil {
			return err
		}
		//删除关联
		err = ResourcesPerformers{}.DeleteByPerformerID(tx, id)
		if err != nil {
			return err
		}
		err = ResourcesDirectors{}.DeleteByPerformerID(tx, id)
		if err != nil {
			return err
		}
		t.DeletePerformerPhoto(info.PerformerBasesID, info.Photo)
		return nil
	})
}

func (t Performer) GetDataListPerformerExpand(performerBasesId string) (*[]PerformerExpand, error) {
	db := core.DBS()
	list, err := models.Performer{}.DataListByPerformerBasesId(db, performerBasesId)
	if err != nil {
		return nil, err
	}
	dataList := make([]PerformerExpand, len(*list))
	for i := range *list {
		photoBase64 := ""
		if (*list)[i].Photo != "" {
			photoBase64, err = t.GetPerformerPhoto((*list)[i].PerformerBasesID, (*list)[i].Photo)
			if err != nil {
				photoBase64 = ""
			}
		}
		dataList[i] = PerformerExpand{PerformerInfo: (*list)[i], PhotoBase64: photoBase64}
	}
	return &dataList, nil
}
