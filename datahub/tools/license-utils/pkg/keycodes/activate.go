package keycodes

import (
	"context"
	"errors"
	"fmt"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

func Activate(filePath string) error {
	// Check if registration file is found
	if !AlamedaUtils.FileExists(filePath) {
		reason := fmt.Sprintf("registration file(%s) is not found", filePath)
		fmt.Println(fmt.Sprintf("[Error]: %s", reason))
		return errors.New(reason)
	}

	// Read registration file
	registrationFile, err := AlamedaUtils.ReadFile(filePath)
	if err != nil {
		reason := fmt.Sprintf("failed to read registration file(%s)", filePath)
		fmt.Println(fmt.Sprintf("[Error]: %s", reason))
		return errors.New(reason)
	}

	// Check if registration file is empty
	if len(registrationFile) == 0 {
		reason := fmt.Sprintf("registration file(%s) is empty", filePath)
		fmt.Println(fmt.Sprintf("[Error]: %s", reason))
		return errors.New(reason)
	}

	// Connect to datahub
	conn, err := grpc.Dial(*datahubAddress, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	client := Keycodes.NewKeycodesServiceClient(conn)

	// Generate request
	in := &Keycodes.ActivateRegistrationDataRequest{
		Data: registrationFile[0],
	}

	// Do API request
	stat, err := client.ActivateRegistrationData(context.Background(), in)
	if err != nil {
		fmt.Println("[Error]: failed to connect to datahub")
		fmt.Println(fmt.Sprintf("[Reason]: %s", err.Error()))
		return err
	}

	// Check API result
	retCode := int32(stat.GetCode())
	if retCode == int32(code.Code_OK) {
		fmt.Println(fmt.Sprintf("[Result]: %s", code.Code_name[retCode]))
	} else {
		fmt.Println(fmt.Sprintf("[Result]: %s", code.Code_name[retCode]))
		fmt.Println(fmt.Sprintf("[Reason]: %s", stat.GetMessage()))
		return errors.New(stat.GetMessage())
	}

	return nil
}
