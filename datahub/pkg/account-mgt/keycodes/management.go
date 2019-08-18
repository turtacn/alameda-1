package keycodes

import (
	"encoding/json"
	"fmt"
	EntityInflux "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb"
	EntityInfluxKeycode "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb/cluster_status"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Events "github.com/containers-ai/alameda/internal/pkg/event-mgt"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"math"
	"strings"
	"time"
)

type KeycodeMgt struct {
	Executor      *KeycodeExecutor
	Status        *KeycodeStatus
	KeycodeStatus string
}

func NewKeycodeMgt() *KeycodeMgt {
	keycodeMgt := KeycodeMgt{}
	keycodeMgt.Executor = NewKeycodeExecutor()
	keycodeMgt.Status = NewKeycodeStatus()
	if KeycodeSummary != nil {
		keycodeMgt.KeycodeStatus = keycodeMgt.GetStatus()
	}
	return &keycodeMgt
}

func (c *KeycodeMgt) AddKeycode(keycode string) error {
	err := c.Executor.AddKeycode(keycode)

	if err != nil {
		scope.Errorf("failed to add keycode(%s)", keycode)
		return err
	}

	c.Refresh(true)

	return nil
}

func (c *KeycodeMgt) DeleteKeycode(keycode string) error {
	err := c.Executor.DeleteKeycode(keycode)

	if err != nil {
		scope.Errorf("failed to delete keycode(%s)", keycode)
		return err
	}

	c.Refresh(true)

	return nil
}

func (c *KeycodeMgt) GetKeycode(keycode string) (*Keycode, error) {
	err := c.Refresh(false)

	if err != nil {
		scope.Errorf("failed to get keycode(%s)", keycode)
		return nil, err
	}

	stripped := strings.Replace(keycode, "-", "", -1)

	for _, keycodeObj := range KeycodeList {
		if keycodeObj.Keycode == stripped {
			return keycodeObj, nil
		}
	}

	return nil, nil
}

func (c *KeycodeMgt) GetKeycodeSummary() (*Keycode, error) {
	err := c.Refresh(false)

	if err != nil {
		scope.Error("failed to get keycode summary")
		return nil, err
	}

	return KeycodeSummary, nil
}

func (c *KeycodeMgt) GetKeycodes(keycodes []string) ([]*Keycode, *Keycode, error) {
	err := c.Refresh(false)

	keycodeList := make([]*Keycode, 0)

	if err != nil {
		scope.Error("failed to get keycodes")
		return nil, nil, err
	}

	for _, keycode := range keycodes {
		stripped := strings.Replace(keycode, "-", "", -1)
		for _, keycodeObj := range KeycodeList {
			if keycodeObj.Keycode == stripped {
				keycodeList = append(keycodeList, keycodeObj)
			}
		}
	}

	return keycodeList, KeycodeSummary, nil
}

func (c *KeycodeMgt) GetAllKeycodes() ([]*Keycode, *Keycode, error) {
	err := c.Refresh(false)

	if err != nil {
		scope.Error("failed to get all keycodes")
		return make([]*Keycode, 0), nil, err
	}

	return KeycodeList, KeycodeSummary, nil
}

func (c *KeycodeMgt) GetRegistrationData() (string, error) {
	registrationData, err := c.Executor.GetRegistrationData()

	if err != nil {
		scope.Error("failed to get registration data")
		return "", err
	}

	return registrationData, nil
}

func (c *KeycodeMgt) PutSignatureData(signatureData string) error {
	err := c.Executor.PutSignatureData(signatureData)

	if err != nil {
		return err
	}

	c.Refresh(true)

	return nil
}

func (c *KeycodeMgt) PutSignatureDataFile(filePath string) error {
	err := c.Executor.PutSignatureDataFile(filePath)

	if err != nil {
		return err
	}

	c.Refresh(true)

	return nil
}

func (c *KeycodeMgt) Refresh(force bool) error {
	tm := time.Now()
	tmUnix := tm.Unix()
	refreshed := false
	keycode := "N/A"

	if (force == true) || (int64(math.Abs(float64(tmUnix-KeycodeTimestamp))) >= KeycodeDuration) {
		keycodeList, keycodeSummary, err := c.Executor.GetAllKeycodes()
		if err != nil {
			scope.Error("failed to refresh keycodes information")
			return err
		}

		KeycodeTimestamp = tmUnix
		KeycodeList = keycodeList
		KeycodeSummary = keycodeSummary
		KeycodeTM = tm
		refreshed = true
	}

	if len(KeycodeList) > 0 {
		// log the first keycode in KeycodeList
		keycode = KeycodeList[0].Keycode
	}
	if force == false {
		if refreshed == true {
			scope.Infof("keycode cache data refreshed, keycode: %s", keycode)
		} else {
			scope.Infof("cached keycode (@ %s): %s", KeycodeTM.Format(time.RFC3339), keycode)
		}
	} else {
		scope.Infof("keycode cache data refreshed for CUD OP, keycode: %s", keycode)
	}

	if c.KeycodeStatus != c.GetStatus() {
		KeycodeLicenseStatus = c.GetStatus()
		c.KeycodeStatus = c.GetStatus()

		// Update InfluxDB and post event
		switch KeycodeLicenseStatus {
		case KeycodeStatusNoKeycode:
			c.writeInfluxEntry("N/A", KeycodeStatusNoKeycode)
			c.deleteInfluxEntry("Summary")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_ERROR, fmt.Sprintf("Keycode state is %s", KeycodeSummary.LicenseState))
		case KeycodeStatusInvalid:
			c.writeInfluxEntry("Summary", KeycodeStatusInvalid)
			c.deleteInfluxEntry("N/A")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_ERROR, fmt.Sprintf("Keycode state is %s", KeycodeSummary.LicenseState))
		case KeycodeStatusExpired:
			c.writeInfluxEntry("Summary", KeycodeStatusExpired)
			c.deleteInfluxEntry("N/A")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_ERROR, fmt.Sprintf("Keycode state is %s", KeycodeSummary.LicenseState))
		case KeycodeStatusNotActivated:
			c.writeInfluxEntry("Summary", KeycodeStatusNotActivated)
			c.deleteInfluxEntry("N/A")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_INFO, fmt.Sprintf("Keycode state is %s", KeycodeSummary.LicenseState))
		case KeycodeStatusValid:
			c.writeInfluxEntry("Summary", KeycodeStatusValid)
			c.deleteInfluxEntry("N/A")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_INFO, fmt.Sprintf("Keycode state is %s", KeycodeSummary.LicenseState))
		default:
			c.writeInfluxEntry("Summary", KeycodeStatusUnknown)
			c.deleteInfluxEntry("N/A")
			c.postEvent(DatahubV1alpha1.EventLevel_EVENT_LEVEL_ERROR, fmt.Sprintf("Keycode state is %s", KeycodeStatusUnknown))
		}
	}

	return nil
}

func (c *KeycodeMgt) GetStatus() string {
	return c.Status.GetStatus()
}

func (c *KeycodeMgt) IsExpired() bool {
	summary, err := c.GetKeycodeSummary()

	if err != nil {
		scope.Error("failed to check if keycode is expired")
		return false
	}

	if summary.LicenseState == "Valid" {
		return false
	}

	return true
}

func (c *KeycodeMgt) postEvent(level DatahubV1alpha1.EventLevel, message string) {
	if level == DatahubV1alpha1.EventLevel_EVENT_LEVEL_INFO {
		scope.Info(message)
	} else {
		scope.Error(message)
	}

	request := &DatahubV1alpha1.CreateEventsRequest{}
	request.Events = append(request.Events, c.generateEvent(level, message))
	Events.PostEvents(request)
}

func (c *KeycodeMgt) writeInfluxEntry(keycode, status string) error {
	points := make([]*InfluxClient.Point, 0)
	client := InternalInflux.NewClient(InfluxConfig)

	tags := map[string]string{
		EntityInfluxKeycode.Keycode: keycode,
	}

	jsonStr, _ := json.Marshal(KeycodeSummary)
	fields := map[string]interface{}{
		EntityInfluxKeycode.KeycodeStatus:          status,
		EntityInfluxKeycode.KeycodeType:            KeycodeSummary.KeycodeType,
		EntityInfluxKeycode.KeycodeState:           KeycodeSummary.LicenseState,
		EntityInfluxKeycode.KeycodeRegistered:      KeycodeSummary.Registered,
		EntityInfluxKeycode.KeycodeExpireTimestamp: KeycodeSummary.ExpireTimestamp,
		EntityInfluxKeycode.KeycodeRawdata:         string(jsonStr[:]),
	}

	pt, err := InfluxClient.NewPoint(string(EntityInfluxKeycode.KeycodeMeasurement), tags, fields, time.Unix(0, 0))
	if err != nil {
		scope.Error(err.Error())
	}
	points = append(points, pt)

	err = client.WritePoints(points, InfluxClient.BatchPointsConfig{
		Database: string(EntityInflux.ClusterStatus),
	})

	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (c *KeycodeMgt) deleteInfluxEntry(keycode string) error {
	if keycode != "" {
		client := InternalInflux.NewClient(InfluxConfig)

		cmd := fmt.Sprintf("DROP SERIES FROM %s WHERE \"%s\"='%s'", EntityInfluxKeycode.KeycodeMeasurement, EntityInfluxKeycode.Keycode, keycode)
		scope.Debugf("delete keycode in influxdb command: %s", cmd)
		_, err := client.QueryDB(cmd, string(EntityInflux.ClusterStatus))
		if err != nil {
			scope.Errorf(err.Error())
			return nil
		}
	}
	return nil
}

func (c *KeycodeMgt) generateEvent(level DatahubV1alpha1.EventLevel, message string) *DatahubV1alpha1.Event {
	event := &DatahubV1alpha1.Event{
		Time:    &timestamp.Timestamp{Seconds: time.Now().Unix()},
		Id:      AlamedaUtils.GenerateUUID(),
		Version: DatahubV1alpha1.EventVersion_EVENT_VERSION_V1,
		Level:   level,
		Message: message,
	}
	return event
}
