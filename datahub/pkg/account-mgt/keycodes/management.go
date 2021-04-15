package keycodes

import (
	"encoding/json"
	"fmt"
	EntityInflux "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb"
	EntityInfluxKeycode "github.com/containers-ai/alameda/internal/pkg/database/entity/influxdb/cluster_status"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	//DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	InfluxClient "github.com/influxdata/influxdb/client/v2"
	"strings"
	"time"
)

type KeycodeMgt struct {
	Executor      *KeycodeExecutor
	Status        *KeycodeStatusObject
	KeycodeStatus int
}

func NewKeycodeMgt() *KeycodeMgt {
	keycodeMgt := KeycodeMgt{}
	keycodeMgt.Executor = NewKeycodeExecutor()
	keycodeMgt.Status = NewKeycodeStatusObject()
	if KeycodeSummary != nil {
		keycodeMgt.KeycodeStatus = keycodeMgt.GetStatus()
	}
	return &keycodeMgt
}

func (c *KeycodeMgt) AddKeycode(keycode string) error {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.Executor.AddKeycode(keycode)

	if err != nil {
		scope.Errorf("failed to add keycode(%s)", keycode)
		return err
	}

	c.refresh(true)

	return nil
}

func (c *KeycodeMgt) DeleteKeycode(keycode string) error {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.Executor.DeleteKeycode(keycode)

	if err != nil {
		scope.Errorf("failed to delete keycode(%s)", keycode)
		return err
	}

	c.refresh(true)

	return nil
}

func (c *KeycodeMgt) GetKeycode(keycode string) (*Keycode, error) {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.refresh(false)

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
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.refresh(false)

	if err != nil {
		scope.Error("failed to get keycode summary")
		return nil, err
	}

	return KeycodeSummary, nil
}

func (c *KeycodeMgt) GetKeycodes(keycodes []string) ([]*Keycode, *Keycode, error) {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.refresh(false)

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
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.refresh(false)

	if err != nil {
		scope.Error("failed to get all keycodes")
		return make([]*Keycode, 0), nil, err
	}

	return KeycodeList, KeycodeSummary, nil
}

func (c *KeycodeMgt) GetRegistrationData() (string, error) {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	registrationData, err := c.Executor.GetRegistrationData()

	if err != nil {
		scope.Error("failed to get registration data")
		return "", err
	}

	return registrationData, nil
}

func (c *KeycodeMgt) PutSignatureData(signatureData string) error {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.Executor.PutSignatureData(signatureData)

	if err != nil {
		return err
	}

	c.refresh(true)

	return nil
}

func (c *KeycodeMgt) PutSignatureDataFile(filePath string) error {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	err := c.Executor.PutSignatureDataFile(filePath)

	if err != nil {
		return err
	}

	c.refresh(true)

	return nil
}

func (c *KeycodeMgt) Refresh(force bool) error {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	return c.refresh(force)
}

func (c *KeycodeMgt) IsValid() bool {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()

	return true
}

func (c *KeycodeMgt) IsExpired() bool {
	KeycodeMutex.Lock()
	defer KeycodeMutex.Unlock()
	return false
}

// NOTE: DO Refresh() before GetStatus() if necessary
func (c *KeycodeMgt) GetStatus() int {
	return c.Status.GetStatus()
}

func (c *KeycodeMgt) refresh(force bool) error { // NOTE: DO GET KeycodeMutex lock before using this function
	return nil
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
