package client

import (
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// IsResponseStatusOK returns true when s.Code equals Code_OK, or return false with error
func IsResponseStatusOK(s *status.Status) (bool, error) {

	if s != nil && s.Code == int32(code.Code_OK) {
		return true, nil
	}

	var err error
	if s == nil {
		err = errors.New("status nil")
	} else {
		err = errors.Errorf("statusCode: %d, message: %s", s.Code, s.Message)
	}
	return false, err
}
