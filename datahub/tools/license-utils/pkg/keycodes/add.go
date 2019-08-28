package keycodes

import (
	"context"
	"errors"
	"fmt"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

func AddKeycode(keycode string) error {
	// Connect to datahub
	conn, err := grpc.Dial(*datahubAddress, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	client := Keycodes.NewKeycodesServiceClient(conn)

	// Generate request
	in := &Keycodes.AddKeycodeRequest{
		Keycode: keycode,
	}

	// Do API request
	response, err := client.AddKeycode(context.Background(), in)
	if err != nil {
		fmt.Println("[Error]: failed to connect to datahub")
		fmt.Println(fmt.Sprintf("[Reason]: %s", err.Error()))
		return err
	}

	// Check API result
	retCode := int32(response.GetStatus().GetCode())
	if retCode == int32(code.Code_OK) {
		fmt.Println(fmt.Sprintf("[Result]: %s", code.Code_name[retCode]))
		fmt.Println("[Keycode]")
		PrintKeycode(response.GetKeycode())
	} else {
		fmt.Println(fmt.Sprintf("[Result]: %s", code.Code_name[retCode]))
		fmt.Println(fmt.Sprintf("[Reason]: %s", response.GetStatus().GetMessage()))
		return errors.New(response.GetStatus().GetMessage())
	}

	return nil
}
