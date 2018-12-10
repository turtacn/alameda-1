package grpc

import (
	"flag"
	"os"
)

func GetAIServiceAddress() string {
	aiServer := os.Getenv("ALAMEDA_AI_SERVER_ADDRESS")
	if len(aiServer) == 0 {
		return "alameda-ai.alameda.svc.cluster.local:50051"
	}
	return aiServer
}

func GetServerPort() int {
	return flag.Lookup("server-port").Value.(flag.Getter).Get().(int)
}
