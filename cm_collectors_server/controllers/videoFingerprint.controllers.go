package controllers

import (
	"cm_collectors_server/processors"
	"cm_collectors_server/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VideoFingerprint struct{}

func (VideoFingerprint) Stats(c *gin.Context) {
	filesBasesID := c.Query("files_bases_id")
	stats, err := processors.VideoFingerprint{}.Stats(filesBasesID)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(stats, c)
}

func (VideoFingerprint) QueryDuplicates(c *gin.Context) {
	filesBasesID := c.Query("files_bases_id")
	matchMode := c.DefaultQuery("match_mode", "full")
	thresholdStr := c.DefaultQuery("threshold", "8")
	durationFirstStr := c.DefaultQuery("duration_first", "true")
	durationToleranceStr := c.DefaultQuery("duration_tolerance", "3")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	threshold, _ := strconv.Atoi(thresholdStr)
	durationFirst, _ := strconv.ParseBool(durationFirstStr)
	durationTolerance, _ := strconv.ParseFloat(durationToleranceStr, 64)
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 30 {
		threshold = 30
	}
	if durationTolerance < 0 {
		durationTolerance = 0
	}
	if !durationFirst {
		durationTolerance = 0
	}

	groups, total, err := processors.VideoFingerprint{}.QueryDuplicates(filesBasesID, matchMode, threshold, durationTolerance, page, limit)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(gin.H{
		"dataList": groups,
		"total":    total,
	}, c)
}

func (VideoFingerprint) TriggerCompute(c *gin.Context) {
	filesBasesID := c.Query("files_bases_id")
	batchSizeStr := c.DefaultQuery("batch_size", "50")
	batchSize, _ := strconv.Atoi(batchSizeStr)
	if batchSize < 1 || batchSize > 500 {
		batchSize = 50
	}

	vfP := processors.VideoFingerprint{}
	if err := vfP.StartComputePendingFingerprintsAsync(filesBasesID, batchSize); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(true, c)
}

func (VideoFingerprint) ReScan(c *gin.Context) {
	filesBasesID := c.Query("files_bases_id")
	count, err := processors.VideoFingerprint{}.ReScanMissingFingerprintsByFilesBasesID(filesBasesID)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(gin.H{"added": count}, c)
}

func (VideoFingerprint) ResetAll(c *gin.Context) {
	err := processors.VideoFingerprint{}.ResetAllFingerprints()
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}

func (VideoFingerprint) ResetFailed(c *gin.Context) {
	filesBasesID := c.Query("files_bases_id")
	count, err := processors.VideoFingerprint{}.ResetFailed(filesBasesID)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(gin.H{"reset": count}, c)
}

func (VideoFingerprint) TaskStatus(c *gin.Context) {
	response.OkWithData(processors.VideoFingerprint{}.TaskStatus(), c)
}

type ReqVideoFingerprintDeleteDramaSeries struct {
	DramaSeriesIDs []string `json:"drama_series_ids" binding:"required"`
	DeleteFile     bool     `json:"delete_file"`
}

func (VideoFingerprint) DeleteDramaSeries(c *gin.Context) {
	var par ReqVideoFingerprintDeleteDramaSeries
	if err := ParameterHandleShouldBindJSON(c, &par); err != nil {
		return
	}
	err := processors.VideoFingerprint{}.DeleteDramaSeries(par.DramaSeriesIDs, par.DeleteFile)
	if err := ResError(c, err); err != nil {
		return
	}
	response.OkWithData(true, c)
}
