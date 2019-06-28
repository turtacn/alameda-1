package utils

import (
	"encoding/json"
	Common "github.com/containers-ai/api/common"
	"os"
)

// InterfaceToString encodes interface to string
func InterfaceToString(data interface{}) string {
	if configBin, err := json.Marshal(data); err != nil {
		return ""
	} else {
		return string(configBin)
	}
}

// GetRunningNamespace retrieves value from env NAMESPACE_NAME
func GetRunningNamespace() string {
	return os.Getenv("NAMESPACE_NAME")
}

// GetRunningPodName retrieves value from env POD_NAME
func GetRunningPodName() string {
	return os.Getenv("POD_NAME")
}

func RawdataRead2Write(readRawdata []*Common.ReadRawdata) []*Common.WriteRawdata {
	writeRawdata := make([]*Common.WriteRawdata, 0)

	for _, rRawdata := range readRawdata {
		wRawdata := Common.WriteRawdata{}

		wRawdata.Database = rRawdata.GetQuery().GetDatabase()
		wRawdata.Table    = rRawdata.GetQuery().GetTable()
		for _, column := range rRawdata.GetColumns() {
			wRawdata.Columns = append(wRawdata.Columns, column)
		}
		for _, group := range rRawdata.GetGroups() {
			for _, row := range group.GetRows() {
				wRawdata.Rows = append(wRawdata.Rows, row)
			}
		}

		writeRawdata = append(writeRawdata, &wRawdata)
	}

	return writeRawdata
}
