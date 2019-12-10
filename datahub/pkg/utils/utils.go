package utils

import (
	"encoding/csv"
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes/timestamp"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	utilsScope = Log.RegisterScope("utils", "utils", 0)
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
func GetSampleInstance(timeObj, endTimeObj *time.Time, numVal string) *ApiCommon.Sample {
	seconds := timeObj.Unix()
	endSeconds := endTimeObj.Unix()
	if timeObj != nil && endTimeObj != nil {
		return &ApiCommon.Sample{
			Time: &timestamp.Timestamp{
				Seconds: seconds,
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endSeconds,
			},
			NumValue: numVal,
		}
	}
	return &ApiCommon.Sample{
		NumValue: numVal,
	}
}

// GetTimeIdxFromColumns get index of time column
func GetTimeIdxFromColumns(columns []string) int {
	for idx, column := range columns {
		if column == RepoInflux.Time {
			return idx
		}
	}
	return 0
}

// GetTimeIdxFromColumns get index of end_time column
func GetEndTimeIdxFromColumns(columns []string) int {
	for idx, column := range columns {
		if column == RepoInflux.EndTime {
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

// StringToFloat64 parse str to float64
func StringToFloat64(str string) (float64, error) {
	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return val, err
	} else {
		return 0, err
	}
}

func ReadCSV(file string) (map[string][]string, error) {
	retMap := map[string][]string{}

	csvFile, err := os.Open(file)
	if err != nil {
		return retMap, nil
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	rows, err := csvReader.ReadAll()

	for _, row := range rows {
		podName := row[0]
		retMap[podName] = make([]string, 0)
		for _, data := range row[1:] {
			retMap[podName] = append(retMap[podName], data)
		}
	}

	return retMap, nil
}

func SliceContains(sliceType interface{}, item interface{}) bool {
	slice := reflect.ValueOf(sliceType)

	if slice.Kind() != reflect.Array && slice.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < slice.Len(); i++ {
		if slice.Index(i).Interface() == item {
			return true
		}
	}

	return false
}
