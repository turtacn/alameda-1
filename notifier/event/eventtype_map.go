package event

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

func EventTypeIntToYamlKeyMap(enumInt int32) string {
	return viper.GetString(fmt.Sprintf("eventType.%v", enumInt))
}

func EventTypeYamlKeyToIntMap(enumKey string) int32 {
	theMap := viper.GetStringMap("eventType")
	for key, val := range theMap {
		if enumKey == val {
			if result, err := strconv.Atoi(key); err != nil {
				return 0
			} else {
				return int32(result)
			}
		}
	}
	return 0
}

func IsEventTypeYamlKeySupported(enumKey string) bool {
	theMap := viper.GetStringMap("eventType")
	for _, val := range theMap {
		if enumKey == val {
			return true
		}
	}
	return false
}

func ListEventTypeYamlKey() []string {
	keys := []string{}
	theMap := viper.GetStringMap("eventType")
	for _, val := range theMap {
		keys = append(keys, val.(string))
	}
	return keys
}
