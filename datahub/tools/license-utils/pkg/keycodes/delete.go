package keycodes

import (
	"context"
	"errors"
	"fmt"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

func DeleteKeycode(keycode string) error {
	// Connect to datahub
	conn, err := grpc.Dial(*datahubAddress, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	client := Keycodes.NewKeycodesServiceClient(conn)

	// Generate request
	in := &Keycodes.DeleteKeycodeRequest{
		Keycode: keycode,
	}

	// Do API request
	stat, err := client.DeleteKeycode(context.Background(), in)
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
