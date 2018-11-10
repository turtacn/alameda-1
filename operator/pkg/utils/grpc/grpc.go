package grpc

import (
	"flag"
)

func GetAIServiceAddress() string {
	return flag.Lookup("ai-server").Value.(flag.Getter).Get().(string)
}
