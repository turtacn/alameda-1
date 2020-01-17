package dispatcher

import (
	"fmt"
	"sync"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/consts"
	"github.com/containers-ai/alameda/pkg/utils"
	"github.com/spf13/viper"
)

type ModelMapper struct {
	modelMap     map[string]*modelInfo
	modelTimeout int64
	lock         *sync.Mutex
}

func NewModelMapper() *ModelMapper {
	modelMap := map[string]*modelInfo{}
	return &ModelMapper{
		modelMap:     modelMap,
		modelTimeout: viper.GetInt64("model.timeout"),
		lock:         &sync.Mutex{},
	}
}

func (mm *ModelMapper) AddModelInfo(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	uniKey := mm.getUniqueName(clusterID, predictUnitType,
		granularity, metricType, extraInfo)
	scope.Debugf("before (AddModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
	scope.Debugf("model added to mapper: %s", uniKey)
	mm.modelMap[uniKey] = &modelInfo{
		Timestamp: time.Now().Unix(),
	}
	scope.Debugf("after (AddModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
}

func (mm *ModelMapper) RemoveModelInfo(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	uniKey := mm.getUniqueName(clusterID, predictUnitType,
		granularity, metricType, extraInfo)
	scope.Debugf("before (RemoveModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
	scope.Debugf("model removed from mapper: %s", uniKey)
	if _, ok := mm.modelMap[uniKey]; ok {
		delete(mm.modelMap, uniKey)
	}
	scope.Debugf("after (RemoveModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
}

func (mm *ModelMapper) GetModelInfo(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) *modelInfo {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	uniKey := mm.getUniqueName(clusterID, predictUnitType,
		granularity, metricType, extraInfo)
	val, _ := mm.modelMap[uniKey]
	return val
}

func (mm *ModelMapper) IsModelTimeout(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) bool {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	scope.Debugf("current mapper status: %s", utils.InterfaceToString(mm.modelMap))
	isTimeout := true
	uniKey := mm.getUniqueName(clusterID, predictUnitType,
		granularity, metricType, extraInfo)
	if oldMInfo, ok := mm.modelMap[uniKey]; ok {
		isTimeout = time.Now().Unix()-oldMInfo.Timestamp > mm.modelTimeout
	} else {
		isTimeout = true
	}
	scope.Debugf("model timeout check from mapper %s is %t",
		uniKey, isTimeout)
	return isTimeout
}

func (mm *ModelMapper) IsModeling(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) bool {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	scope.Debugf("current mapper status: %s", utils.InterfaceToString(mm.modelMap))
	uniKey := mm.getUniqueName(clusterID, predictUnitType,
		granularity, metricType, extraInfo)
	isModeling := false
	_, ok := mm.modelMap[uniKey]
	isModeling = ok
	scope.Debugf("is model check from mapper %s is %t",
		uniKey, isModeling)
	return isModeling
}

func (mm *ModelMapper) getUniqueName(clusterID string, predictUnitType string,
	granularity string, metricType string, extraInfo map[string]string) string {
	if predictUnitType == consts.UnitTypeNode {
		return fmt.Sprintf("%s/%s/%s/%s/%s", predictUnitType, clusterID, extraInfo["name"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypePod {
		return fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", predictUnitType, clusterID,
			extraInfo["namespace"], extraInfo["name"], extraInfo["containerName"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypeGPU {
		return fmt.Sprintf("%s/%s/%s/%s/%s/%s", predictUnitType, clusterID,
			extraInfo["host"], extraInfo["minorNumber"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypeNamespace {
		return fmt.Sprintf("%s/%s/%s/%s/%s", predictUnitType, clusterID, extraInfo["name"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypeController {
		return fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", predictUnitType, extraInfo["kind"], clusterID,
			extraInfo["namespace"], extraInfo["name"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypeApplication {
		return fmt.Sprintf("%s/%s/%s/%s/%s/%s", predictUnitType, clusterID,
			extraInfo["namespace"], extraInfo["name"], granularity, metricType)
	} else if predictUnitType == consts.UnitTypeCluster {
		return fmt.Sprintf("%s/%s/%s/%s", predictUnitType, extraInfo["name"], granularity, metricType)
	}
	return ""
}
