package utils

import (
	"strconv"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
)

var (
	utilsScope = logUtil.RegisterScope("utils", "utils", 0)
)

type StringStringMap map[string]string

func (m StringStringMap) ReplaceKeys(old, new []string) StringStringMap {

	for i, oldKey := range old {
		if v, exist := m[oldKey]; exist {
			newKey := new[i]
			delete(m, oldKey)
			m[newKey] = v
		}
	}

	return m
}

func ParseTime(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)

	return t, err
}

// NanoSecondToSecond translate nano seconds to seconds
func NanoSecondToSecond(nanosecond int64) int64 {
	return nanosecond / 1000000000
}

// GetSampleInstance get Sample instance
func GetSampleInstance(timeObj *time.Time, numVal string) *datahub_v1alpha1.Sample {
	seconds := timeObj.Unix()
	if timeObj != nil {
		return &datahub_v1alpha1.Sample{
			Time: &timestamp.Timestamp{
				Seconds: seconds,
			},
			NumValue: numVal,
		}
	}
	return &datahub_v1alpha1.Sample{
		NumValue: numVal,
	}
}

// GetTimeIdxFromColumns get index of time column
func GetTimeIdxFromColumns(columns []string) int {
	for idx, column := range columns {
		if column == influxdb.Time {
			return idx
		}
	}
	return 0
}

// TimeStampToNanoSecond get nano seconds from timestamp object
func TimeStampToNanoSecond(timestamp *timestamp.Timestamp) int64 {
	return timestamp.GetSeconds()*1000000000 + int64(timestamp.GetNanos())
}

// StringToInt64 parse str to int64
func StringToInt64(str string) (int64, error) {

	if val, err := strconv.ParseInt(str, 10, 64); err == nil {
		return val, err
	}

	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return int64(val), err
	} else {
		return 0, err
	}
}
