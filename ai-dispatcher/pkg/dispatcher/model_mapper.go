package dispatcher

import (
	"fmt"
	"time"

	"github.com/containers-ai/alameda/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
)

type ModelMapper struct {
	modelMap     map[string]map[string]map[string]*modelInfo
	modelTimeout int64
}

func NewModelMapper(predictUnitTypes []string, granularities []string) *ModelMapper {
	modelMap := map[string]map[string]map[string]*modelInfo{}
	for _, pdUnit := range predictUnitTypes {
		for _, gnu := range granularities {
			if _, ok := modelMap[pdUnit]; !ok {
				modelMap[pdUnit] = map[string]map[string]*modelInfo{}
			}
			modelMap[pdUnit][gnu] = map[string]*modelInfo{}
		}
	}
	return &ModelMapper{
		modelMap:     modelMap,
		modelTimeout: viper.GetInt64("model.timeout"),
	}
}

type namespacedName struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type container struct {
	Name         string                        `json:"name"`
	ModelMetrics []datahub_v1alpha1.MetricType `json:"modelMetrics"`
}
type podModel struct {
	NamespacedName *namespacedName `json:"namespaced_name,omitempty"`
	Containers     []*container    `json:"containers,omitempty"`
}

type nodeModel struct {
	Name string `json:"name,omitempty"`
}
type gpuModel struct {
	Host        string `json:"host,omitempty"`
	MinorNumber string `json:"minor_number,omitempty"`
}

func (mm *ModelMapper) AddModelInfo(predictUnitType string,
	granularity string, mInfo *modelInfo) {
	scope.Debugf("before (AddModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
	scope.Debugf("model added to mapper (unit type: %s, granularity: %s): %s",
		predictUnitType, granularity, utils.InterfaceToString(mInfo))
	if _, ok := mm.modelMap[predictUnitType]; !ok {
		mm.modelMap[predictUnitType] = map[string]map[string]*modelInfo{}
	}
	if _, ok := mm.modelMap[predictUnitType][granularity]; !ok {
		mm.modelMap[predictUnitType][granularity] = map[string]*modelInfo{}
	}
	mm.modelMap[predictUnitType][granularity][mm.getUniqueName(predictUnitType, mInfo)] = mInfo
	scope.Debugf("after (AddModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
}

func (mm *ModelMapper) RemoveModelInfo(predictUnitType string,
	granularity string, modelId string) {
	scope.Debugf("before (RemoveModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
	scope.Debugf("model removed from mapper (unit type: %s, granularity: %s): %s",
		predictUnitType, granularity, modelId)
	if _, ok := mm.modelMap[predictUnitType]; !ok {
		scope.Debugf("after (RemoveModelInfo) current mapper status: %s",
			utils.InterfaceToString(mm.modelMap))
		return
	}
	if _, ok := mm.modelMap[predictUnitType][granularity]; !ok {
		scope.Debugf("after (RemoveModelInfo) current mapper status: %s",
			utils.InterfaceToString(mm.modelMap))
		return
	}
	if _, ok := mm.modelMap[predictUnitType][granularity][modelId]; ok {
		delete(mm.modelMap[predictUnitType][granularity], modelId)
	}
	scope.Debugf("after (RemoveModelInfo) current mapper status: %s",
		utils.InterfaceToString(mm.modelMap))
}

func (mm *ModelMapper) GetModelInfo(predictUnitType string,
	granularity string, modelId string) *modelInfo {
	if _, ok := mm.modelMap[predictUnitType]; !ok {
		return nil
	}
	if _, ok := mm.modelMap[predictUnitType][granularity]; !ok {
		return nil
	}
	val, _ := mm.modelMap[predictUnitType][granularity][modelId]
	return val
}

func (mm *ModelMapper) IsModelTimeout(predictUnitType string,
	granularity string, mInfo *modelInfo) bool {
	scope.Debugf("current mapper status: %s", utils.InterfaceToString(mm.modelMap))
	isTimeout := true
	if _, ok := mm.modelMap[predictUnitType]; !ok {
		isTimeout = true
	}
	if _, ok := mm.modelMap[predictUnitType][granularity]; !ok {
		isTimeout = true
	}
	idName := mm.getUniqueName(predictUnitType, mInfo)
	if oldMInfo, ok := mm.modelMap[predictUnitType][granularity][idName]; ok {
		isTimeout = time.Now().Unix()-oldMInfo.GetTimeStamp() > mm.modelTimeout
	} else {
		isTimeout = true
	}
	scope.Debugf("model timeout check from mapper (unit type: %s, granularity: %s) is %t: %s",
		predictUnitType, granularity, isTimeout, utils.InterfaceToString(mInfo))
	return isTimeout
}

func (mm *ModelMapper) IsModeling(predictUnitType string,
	granularity string, mInfo *modelInfo) bool {
	scope.Debugf("current mapper status: %s", utils.InterfaceToString(mm.modelMap))
	isModeling := false
	if _, ok := mm.modelMap[predictUnitType]; !ok {
		isModeling = false
	}
	if _, ok := mm.modelMap[predictUnitType][granularity]; !ok {
		isModeling = false
	}
	idName := mm.getUniqueName(predictUnitType, mInfo)
	_, ok := mm.modelMap[predictUnitType][granularity][idName]
	isModeling = ok
	scope.Debugf("is model check from mapper (unit type: %s, granularity: %s) is %t: %s",
		predictUnitType, granularity, isModeling, utils.InterfaceToString(mInfo))
	return isModeling
}

func (mm *ModelMapper) getUniqueName(predictUnitType string,
	modelInfo *modelInfo) string {
	if predictUnitType == UnitTypeNode {
		return modelInfo.Name
	} else if predictUnitType == UnitTypePod {
		return fmt.Sprintf("%s/%s",
			modelInfo.NamespacedName.Namespace, modelInfo.NamespacedName.Name)
	} else if predictUnitType == UnitTypeGPU {
		return fmt.Sprintf("%s/%s",
			modelInfo.Host, modelInfo.MinorNumber)
	}
	return ""
}
