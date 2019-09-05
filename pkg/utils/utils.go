package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	Common "github.com/containers-ai/api/common"
)

// InterfaceToString encodes interface to string
func InterfaceToString(data interface{}) string {
	if configBin, err := json.Marshal(data); err != nil {
		return ""
	} else {
		return string(configBin)
	}
}

func StringToByteArray(str string) []byte {
	var data = []byte(str)
	return data
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
		wRawdata.Table = rRawdata.GetQuery().GetTable()
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

func GenerateUUID() string {
	// generate 32 bits timestamp
	unix32bits := uint32(time.Now().UTC().Unix())

	buff := make([]byte, 12)

	numRead, err := rand.Read(buff)

	if numRead != len(buff) || err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x-%x", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
}

func IsEmailValid(email string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(email)
}
