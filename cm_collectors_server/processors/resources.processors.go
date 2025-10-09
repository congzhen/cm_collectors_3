package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/errorMessage"
	"cm_collectors_server/models"
	"cm_collectors_server/utils"
	"fmt"
	"net/url"
	"os"
	"path"

	"gorm.io/gorm"
)

type Resources struct{}

func (Resources) DataList(par *datatype.ReqParam_ResourcesList) (*[]models.Resources, int64, error) {
	return models.Resources{}.DataList(core.DBS(), par)
}

func (Resources) DataListByIds(ids []string) (*[]models.Resources, error) {
	return models.Resources{}.DataListByIds(core.DBS(), ids)
}

func (Resources) DataListAll(page, limit int) (*[]models.Resources, error) {
	return models.Resources{}.DataListAll(core.DBS(), page, limit)
}

func (Resources) Info(id string) (*models.Resources, error) {
	info, err := models.Resources{}.Info(core.DBS(), id)
	if err == nil && info.ID == "" || err == gorm.ErrRecordNotFound {
		err = errorMessage.Err_Resources_Not_Found
		return info, err
	}
	return info, nil
}

func (Resources) SampleImages(id, imagePath string) ([]string, error) {
	var decodedImagePath = ""
	var err error
	if imagePath != "" {
		decodedImagePath, err = url.QueryUnescape(imagePath)
		if err != nil {
			return []string{}, err
		}
	}

	firstDramaSeries, err := ResourcesDramaSeries{}.FirstInfoByResourcesID(id)
	if err != nil {
		return []string{}, err
	}
	if firstDramaSeries.Src == "" || !utils.FileExists(firstDramaSeries.Src) {
		return []string{}, nil
	}

	fullImagesPath := utils.GetDirPathFromFilePath(firstDramaSeries.Src)
	if decodedImagePath != "" {
		fullImagesPath = path.Join(fullImagesPath, decodedImagePath)
	}
	imagePaths, err := utils.GetFilesByExtensions([]string{fullImagesPath}, utils.FileImageExtensions, false)
	newImagePaths := make([]string, len(imagePaths))
	for i, imagePath := range imagePaths {
		//获取去掉folderPath的文件名
		newImagePaths[i] = utils.TrimBasePath(imagePath, fullImagesPath)
	}
	return newImagePaths, nil
}
func (Resources) SampleImageBytes(id, imagePath string) (string, []byte, error) {
	decodedImagePath, err := url.QueryUnescape(imagePath)
	if err != nil {
		return "", nil, err
	}
	cleanPath := utils.SanitizePath(decodedImagePath)
	firstDramaSeries, err := ResourcesDramaSeries{}.FirstInfoByResourcesID(id)
	if err != nil {
		return "", nil, err
	}
	if firstDramaSeries.Src == "" || !utils.FileExists(firstDramaSeries.Src) {
		return "", nil, fmt.Errorf("图片不存在")
	}
	folderPath := utils.GetDirPathFromFilePath(firstDramaSeries.Src)
	fullImagePath := path.Join(folderPath, cleanPath)
	ext := utils.FileExt(cleanPath)
	imageBytes, err := utils.ReadFile(fullImagePath)
	return ext, imageBytes, err
}

func (t Resources) CreateResource(par *datatype.ReqParam_Resource) (*models.Resources, error) {
	dbs := core.DBS()
	var id string
	err := dbs.Transaction(func(tx *gorm.DB) error {
		var err error
		id, err = t.Create(tx, par)
		if err != nil {
			return err
		}
		return t.SetResources(tx, id, par, true)
	})
	if err != nil {
		return nil, err
	}
	return t.Info(id)
}

func (t Resources) UpdateResource(par *datatype.ReqParam_Resource, setResourcesDramaSeries bool) (*models.Resources, error) {
	dbs := core.DBS()
	var id string
	err := dbs.Transaction(func(tx *gorm.DB) error {
		var err error
		id, err = t.Update(tx, par)
		if err != nil {
			return err
		}
		return t.SetResources(tx, id, par, setResourcesDramaSeries)
	})
	if err != nil {
		return nil, err
	}
	return t.Info(id)
}

func (t Resources) UpdateResourceTag(resourceID string, tags []string) (*models.Resources, error) {
	dbs := core.DBS()
	err := ResourcesTags{}.SetResourcesTag(dbs, resourceID, tags)
	if err != nil {
		return nil, err
	}
	return t.Info(resourceID)
}

func (t Resources) DeleteResource(resourceId string) error {
	info, err := t.Info(resourceId)
	if err != nil {
		return err
	}
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		err := ResourcesDirectors{}.DeleteByResourcesID(tx, resourceId)
		if err != nil {
			return err
		}
		err = ResourcesPerformers{}.DeleteByResourcesID(tx, resourceId)
		if err != nil {
			return err
		}
		err = ResourcesTags{}.DeleteByResourcesID(tx, resourceId)
		if err != nil {
			return err
		}
		err = ResourcesDramaSeries{}.DeleteByResourcesID(tx, resourceId)
		if err != nil {
			return err
		}
		err = t.DeleteById(tx, resourceId)
		if err != nil {
			return err
		}
		t.DeleteResourcePhoto(info.FilesBasesID, info.CoverPoster)
		return nil
	})
}

func (Resources) SetResources(db *gorm.DB, resourceID string, par *datatype.ReqParam_Resource, setResourcesDramaSeries bool) error {
	return db.Transaction(func(tx *gorm.DB) error {
		err := ResourcesPerformers{}.SetResourcesPerformers(tx, resourceID, par.Performers)
		if err != nil {
			return err
		}
		err = ResourcesDirectors{}.SetResourcesDirectors(tx, resourceID, par.Directors)
		if err != nil {
			return err
		}
		err = ResourcesTags{}.SetResourcesTag(tx, resourceID, par.Tags)
		if err != nil {
			return err
		}
		if setResourcesDramaSeries {
			err = ResourcesDramaSeries{}.SetResourcesDramaSeries(tx, resourceID, par.DramaSeries)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (Resources) UpdateResourcePlay(resourceInfo *models.Resources, lastPlayFile string) error {
	db := core.DBS()
	return db.Transaction(func(tx *gorm.DB) error {
		lastPlayTime := datatype.CustomTime(core.TimeNow())
		resourceModels := models.Resources{
			ID:           resourceInfo.ID,
			Hot:          resourceInfo.Hot + 1,
			LastPlayTime: &lastPlayTime,
			LastPlayFile: lastPlayFile,
		}
		err := resourceModels.Update(tx, &resourceModels, []string{"Hot", "LastPlayTime", "LastPlayFile"})
		if err != nil {
			return err
		}
		tagIds := []string{}
		for _, tag := range resourceInfo.Tags {
			tagIds = append(tagIds, tag.ID)
		}
		err = Tag{}.UpdateHot(tx, tagIds)
		if err != nil {
			return err
		}
		return nil
	})
}

func (Resources) SaveResourcePhoto(par *datatype.ReqParam_Resource) (string, error) {
	if par.PhotoBase64 == "" || par.Resource.FilesBasesID == "" {
		return "", nil // 如果没有图片数据或PerformerBasesID，则不处理
	}
	// 生成唯一的图片名称
	photoName := fmt.Sprintf("%s.jpg", core.GenerateUniqueID())
	filePath := path.Join(core.Config.System.FilePath, "resCoverPoster", par.Resource.FilesBasesID, photoName)

	// 将Base64图片保存为文件
	err := utils.SaveBase64AsImage(par.PhotoBase64, filePath, true)
	if err != nil {
		return "", errorMessage.WrapError(errorMessage.Err_Resources_Save_Photo_Failed, err)
	}
	return photoName, nil
}
func (Resources) DeleteResourcePhoto(filesBasesID, photoName string) error {
	if photoName == "" {
		return nil // 如果没有旧图片，则不处理
	}
	oldFilePath := path.Join(core.Config.System.FilePath, "resCoverPoster", filesBasesID, photoName)
	err := os.Remove(oldFilePath)
	if err != nil && !os.IsNotExist(err) {
		return errorMessage.WrapError(errorMessage.Err_Resources_Delete_Photo_Failed, err)
	}
	return nil
}

func (t Resources) Create(tx *gorm.DB, par *datatype.ReqParam_Resource) (string, error) {
	id := core.GenerateUniqueID()
	coverPoster, err := t.SaveResourcePhoto(par)
	if err != nil {
		return "", err
	}
	var issuingDate datatype.CustomDate
	issuingDate.SetValue(par.Resource.IssuingDate)
	var createdAt datatype.CustomTime
	createdAt.SetValue(core.TimeNow())
	resourcesModels := models.Resources{
		ID:                id,
		FilesBasesID:      par.Resource.FilesBasesID,
		Title:             par.Resource.Title,
		KeyWords:          utils.PinyinInitials(par.Resource.Title),
		IssueNumber:       par.Resource.IssueNumber,
		Mode:              par.Resource.Mode,
		CoverPoster:       coverPoster,
		CoverPosterMode:   par.Resource.CoverPosterMode,
		CoverPosterWidth:  par.Resource.CoverPosterWidth,
		CoverPosterHeight: par.Resource.CoverPosterHeight,
		IssuingDate:       &issuingDate,
		Country:           par.Resource.Country,
		Definition:        par.Resource.Definition,
		Stars:             par.Resource.Stars,
		Abstract:          par.Resource.Abstract,
		CreatedAt:         &createdAt,
		Status:            true,
	}
	if par.Resource.LastScraperUpdateTime != nil {
		resourcesModels.LastScraperUpdateTime = par.Resource.LastScraperUpdateTime
	}
	return id, resourcesModels.Create(tx, &resourcesModels)
}

func (t Resources) Update(db *gorm.DB, par *datatype.ReqParam_Resource) (string, error) {
	id := par.Resource.ID
	if id == "" {
		return "", errorMessage.Err_Resources_ID_Empty
	}
	info, err := t.Info(id)
	if err != nil {
		return "", err
	}
	coverPoster, err := t.SaveResourcePhoto(par)
	if err != nil {
		return "", err
	}
	var issuingDate datatype.CustomDate
	issuingDate.SetValue(par.Resource.IssuingDate)
	resourcesModels := models.Resources{
		ID:                id,
		FilesBasesID:      par.Resource.FilesBasesID,
		Title:             par.Resource.Title,
		KeyWords:          utils.PinyinInitials(par.Resource.Title),
		IssueNumber:       par.Resource.IssueNumber,
		Mode:              par.Resource.Mode,
		CoverPosterMode:   par.Resource.CoverPosterMode,
		CoverPosterWidth:  par.Resource.CoverPosterWidth,
		CoverPosterHeight: par.Resource.CoverPosterHeight,
		IssuingDate:       &issuingDate,
		Country:           par.Resource.Country,
		Definition:        par.Resource.Definition,
		Stars:             par.Resource.Stars,
		Abstract:          par.Resource.Abstract,
	}
	fieldsToUpdate := []string{
		"filesBases_id",
		"title",
		"keyWords",
		"issueNumber",
		"mode",
		"coverPosterMode",
		"coverPosterWidth",
		"coverPosterHeight",
		"issuingDate",
		"country",
		"definition",
		"stars",
		"abstract",
	}
	if coverPoster != "" {
		resourcesModels.CoverPoster = coverPoster
		t.DeleteResourcePhoto(info.FilesBasesID, info.CoverPoster)
		fieldsToUpdate = append(fieldsToUpdate, "coverPoster")
	}
	if par.Resource.LastScraperUpdateTime != nil {
		resourcesModels.LastScraperUpdateTime = par.Resource.LastScraperUpdateTime
		fieldsToUpdate = append(fieldsToUpdate, "lastScraperUpdateTime")
	}
	return id, resourcesModels.Update(db, &resourcesModels, fieldsToUpdate)
}

func (t Resources) DeleteById(db *gorm.DB, id string) error {
	return models.Resources{}.DeleteById(db, id)
}
