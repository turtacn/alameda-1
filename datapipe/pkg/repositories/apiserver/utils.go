package apiserver

import (
	"fmt"
    "google.golang.org/genproto/googleapis/rpc/code"
    "google.golang.org/genproto/googleapis/rpc/status"
)

func NeedResendRequest(status *status.Status, err error) bool {
	if err != nil {
		fmt.Print(err)
		return false
	}

	if status.GetCode() == int32(code.Code_UNAUTHENTICATED){
		GetToken(true, 5)
		return true
	}

	return false
}

func CheckResponse(stat *status.Status, err error) (*status.Status, error) {
	// If err is not nil which means there is some issues in connection to API server
	if err != nil {
		fmt.Print(err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	if stat.GetCode() != int32(code.Code_OK) {
		fmt.Print(stat.GetMessage())
		return &status.Status{
			Code:    int32(stat.GetCode()),
			Message: stat.GetMessage(),
		}, nil
	}

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
