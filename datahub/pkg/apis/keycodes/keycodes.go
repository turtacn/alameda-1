package keycodes

import (
	KeycodeMgt "github.com/containers-ai/alameda/datahub/pkg/account-mgt/keycodes"
	DatahubConfig "github.com/containers-ai/alameda/datahub/pkg/config"
	Errors "github.com/containers-ai/alameda/internal/pkg/errors"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/rpc/code"
	"time"
)

var (
	scope = Log.RegisterScope("datahub", "datahub keycodes log", 0)
)

type IError = Errors.InternalError

type ServiceKeycodes struct {
	Config *DatahubConfig.Config
}

func NewService(cfg *DatahubConfig.Config) *ServiceKeycodes {
	service := ServiceKeycodes{}
	service.Config = cfg
	return &service
}

func NewKeycode() *Keycodes.Keycode {
	keycode := Keycodes.Keycode{}
	keycode.Capacity = &Keycodes.Capacity{}
	keycode.Functionality = &Keycodes.Functionality{}
	keycode.Retention = &Keycodes.Retention{}
	keycode.ServiceAgreement = &Keycodes.ServiceAgreement{}
	return &keycode
}

func TransformKeycode(keycode *KeycodeMgt.Keycode) *Keycodes.Keycode {
	keycodeInfo := NewKeycode()
	keycodeInfo.Keycode = keycode.Keycode
	keycodeInfo.KeycodeType = keycode.KeycodeType
	keycodeInfo.KeycodeVersion = int32(keycode.KeycodeVersion)
	keycodeInfo.Registered = keycode.Registered
	keycodeInfo.LicenseState = keycode.LicenseState
	keycodeInfo.Capacity.Users = int32(keycode.Capacity.Users)
	keycodeInfo.Capacity.Hosts = int32(keycode.Capacity.Hosts)
	keycodeInfo.Capacity.Disks = int32(keycode.Capacity.Disks)
	keycodeInfo.Functionality.DiskProphet = keycode.Functionality.Diskprophet
	keycodeInfo.Functionality.Workload = keycode.Functionality.Workload
	keycodeInfo.Retention.ValidMonth = int32(keycode.Retention.ValidMonth)
	keycodeInfo.Retention.Years = int32(keycode.Retention.Years)

	if keycode.ApplyTimestamp == 0 {
		keycodeInfo.ApplyTime = &timestamp.Timestamp{Seconds: 0}
	} else if keycode.ApplyTimestamp == -1 {
		keycodeInfo.ApplyTime = &timestamp.Timestamp{Seconds: time.Date(2039, 12, 31, 0, 0, 0, 0, time.UTC).Unix()}
	} else {
		keycodeInfo.ApplyTime = &timestamp.Timestamp{Seconds: keycode.ApplyTimestamp}
	}

	if keycode.ExpireTimestamp == 0 {
		keycodeInfo.ExpireTime = &timestamp.Timestamp{Seconds: 0}
	} else if keycode.ExpireTimestamp == -1 {
		keycodeInfo.ExpireTime = &timestamp.Timestamp{Seconds: time.Date(2039, 12, 31, 0, 0, 0, 0, time.UTC).Unix()}
	} else {
		keycodeInfo.ExpireTime = &timestamp.Timestamp{Seconds: keycode.ExpireTimestamp}
	}

	return keycodeInfo
}

func TransformKeycodeList(keycodes []*KeycodeMgt.Keycode) []*Keycodes.Keycode {
	keycodeList := make([]*Keycodes.Keycode, 0)

	for _, keycode := range keycodes {
		keycodeList = append(keycodeList, TransformKeycode(keycode))
	}

	return keycodeList
}

func CategorizeKeycodeErrorId(errorId int) int32 {
	switch errorId {
	case Errors.ReasonKeycodeInvalidKeycode:
		return int32(code.Code_INVALID_ARGUMENT)
	}
	return int32(code.Code_INTERNAL)
}
