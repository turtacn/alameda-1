package event

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

func EventLevelIntToYamlKeyMap(enumInt int32) string {
	return viper.GetString(fmt.Sprintf("eventLevel.%v", enumInt))
}

func EventLevelYamlKeyToIntMap(enumKey string) int32 {
	theMap := viper.GetStringMap("eventLevel")
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
